package parser

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
)

const (
	commandPrefix         = "-- +migrate"
	commandUp             = "up"
	commandDown           = "down"
	commandStatementBegin = "statement_begin"
	commandStatementEnd   = "statement_end"
	optionNoTransaction   = "no_transaction"
)

var (
	ErrNoSemicolon         = errors.New("statement must be ended by a semicolon")
	ErrIncompleteCommand   = errors.New("incomplete migration command")
	ErrUnknownCommand      = errors.New("unknown migration command after prefix")
	ErrStatementNotEnded   = errors.New("statement was started but not ended")
	ErrStatementNotStarted = errors.New("statement was ended but not started")
	ErrNoUpDownCommands    = errors.New("no Up and Down commands found during parsing")
)

type ParsedMigration struct {
	UpStatements   []string
	DownStatements []string

	DisableTransactionUp   bool
	DisableTransactionDown bool
}

func endsWithSemicolon(line string) bool {
	scanner := bufio.NewScanner(strings.NewReader(line))
	scanner.Split(bufio.ScanWords)

	prev := ""
	for scanner.Scan() {
		word := scanner.Text()
		if strings.HasPrefix(word, "--") {
			break
		}
		prev = word
	}

	return strings.HasSuffix(prev, ";")
}

type migrationDirection int

const (
	directionNone migrationDirection = iota
	directionUp
	directionDown
)

type migrateCommand struct {
	command string
	options []string
}

func parseCommand(line string) (*migrateCommand, error) {
	fields := strings.Fields(strings.TrimPrefix(line, commandPrefix))
	if len(fields) == 0 {
		return nil, ErrIncompleteCommand
	}

	return &migrateCommand{
		command: fields[0],
		options: fields[1:],
	}, nil
}

func (c *migrateCommand) hasOption(option string) bool {
	for _, o := range c.options {
		if o == option {
			return true
		}
	}

	return false
}

func isCommand(line string) bool {
	return strings.HasPrefix(line, commandPrefix)
}

func isSQLComment(line string) bool {
	return strings.HasPrefix(line, "--")
}

func isEmpty(line string) bool {
	return strings.TrimSpace(line) == ""
}

func ParseMigration(r io.Reader) (*ParsedMigration, error) { //nolint:all
	pm := &ParsedMigration{}

	var buf bytes.Buffer
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	var statementStarted bool
	var statementEnded bool
	currentDirection := directionNone

	for scanner.Scan() {
		line := scanner.Text()

		if isEmpty(line) || (isSQLComment(line) && !isCommand(line)) {
			continue
		}

		if isCommand(line) { //nolint:all
			cmd, err := parseCommand(line)
			if err != nil {
				return nil, err
			}

			switch cmd.command {
			case commandUp:
				if buf.Len() > 0 {
					return nil, ErrNoSemicolon
				}
				currentDirection = directionUp
				if cmd.hasOption(optionNoTransaction) {
					pm.DisableTransactionUp = true
				}

			case commandDown:
				if buf.Len() > 0 {
					return nil, ErrNoSemicolon
				}
				currentDirection = directionDown
				if cmd.hasOption(optionNoTransaction) {
					pm.DisableTransactionDown = true
				}

			case commandStatementBegin:
				if currentDirection != directionNone {
					statementStarted = true
				}

			case commandStatementEnd:
				if !statementStarted {
					return nil, ErrStatementNotStarted
				}
				if currentDirection != directionNone {
					statementEnded = true
					statementStarted = false
				}

			default:
				return nil, ErrUnknownCommand
			}
		}

		if currentDirection == directionNone {
			continue
		}

		if !isCommand(line) {
			if _, err := buf.WriteString(line + "\n"); err != nil {
				return nil, fmt.Errorf("failed to write string to buffer: %w", err)
			}
		}

		if (!statementStarted && endsWithSemicolon(line)) || statementEnded {
			if currentDirection == directionUp {
				pm.UpStatements = append(pm.UpStatements, buf.String())
			} else {
				pm.DownStatements = append(pm.DownStatements, buf.String())
			}

			statementEnded = false
			buf.Reset()
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan strings: %w", err)
	}

	if statementStarted {
		return nil, ErrStatementNotEnded
	}

	if buf.Len() > 0 {
		return nil, ErrNoSemicolon
	}

	if currentDirection == directionNone {
		return nil, ErrNoUpDownCommands
	}

	return pm, nil
}
