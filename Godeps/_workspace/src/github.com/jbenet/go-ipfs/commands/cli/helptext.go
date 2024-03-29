package cli

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"text/template"

	cmds "github.com/maybebtc/interplanetary/Godeps/_workspace/src/github.com/jbenet/go-ipfs/commands"
)

const (
	requiredArg = "<%v>"
	optionalArg = "[<%v>]"
	variadicArg = "%v..."
	optionFlag  = "-%v"
	optionType  = "(%v)"

	whitespace = "\r\n\t "

	indentStr = "    "
)

type helpFields struct {
	Indent      string
	Usage       string
	Path        string
	ArgUsage    string
	Tagline     string
	Arguments   string
	Options     string
	Synopsis    string
	Subcommands string
	Description string
	MoreHelp    bool
}

// TrimNewlines removes extra newlines from fields. This makes aligning
// commands easier. Below, the leading + tralining newlines are removed:
//	Synopsis: `
//	    ipfs config <key>          - Get value of <key>
//	    ipfs config <key> <value>  - Set value of <key> to <value>
//	    ipfs config --show         - Show config file
//	    ipfs config --edit         - Edit config file in $EDITOR
//	`
func (f *helpFields) TrimNewlines() {
	f.Path = strings.Trim(f.Path, "\n")
	f.ArgUsage = strings.Trim(f.ArgUsage, "\n")
	f.Tagline = strings.Trim(f.Tagline, "\n")
	f.Arguments = strings.Trim(f.Arguments, "\n")
	f.Options = strings.Trim(f.Options, "\n")
	f.Synopsis = strings.Trim(f.Synopsis, "\n")
	f.Subcommands = strings.Trim(f.Subcommands, "\n")
	f.Description = strings.Trim(f.Description, "\n")
}

// Indent adds whitespace the lines of fields.
func (f *helpFields) IndentAll() {
	indent := func(s string) string {
		if s == "" {
			return s
		}
		return indentString(s, indentStr)
	}

	f.Arguments = indent(f.Arguments)
	f.Options = indent(f.Options)
	f.Synopsis = indent(f.Synopsis)
	f.Subcommands = indent(f.Subcommands)
	f.Description = indent(f.Description)
}

const usageFormat = "{{if .Usage}}{{.Usage}}{{else}}{{.Path}}{{if .ArgUsage}} {{.ArgUsage}}{{end}} - {{.Tagline}}{{end}}"

const longHelpFormat = `
{{.Indent}}{{template "usage" .}}

{{if .Arguments}}ARGUMENTS:

{{.Arguments}}

{{end}}{{if .Options}}OPTIONS:

{{.Options}}

{{end}}{{if .Subcommands}}SUBCOMMANDS:

{{.Subcommands}}

{{.Indent}}Use '{{.Path}} <subcmd> --help' for more information about each command.

{{end}}{{if .Description}}DESCRIPTION:

{{.Description}}

{{end}}
`
const shortHelpFormat = `USAGE:

{{.Indent}}{{template "usage" .}}
{{if .Synopsis}}
{{.Synopsis}}
{{end}}{{if .Description}}
{{.Description}}
{{end}}
{{if .MoreHelp}}Use '{{.Path}} --help' for more information about this command.
{{end}}
`

var usageTemplate *template.Template
var longHelpTemplate *template.Template
var shortHelpTemplate *template.Template

func init() {
	usageTemplate = template.Must(template.New("usage").Parse(usageFormat))
	longHelpTemplate = template.Must(usageTemplate.New("longHelp").Parse(longHelpFormat))
	shortHelpTemplate = template.Must(usageTemplate.New("shortHelp").Parse(shortHelpFormat))
}

// LongHelp returns a formatted CLI helptext string, generated for the given command
func LongHelp(rootName string, root *cmds.Command, path []string, out io.Writer) error {
	cmd, err := root.Get(path)
	if err != nil {
		return err
	}

	pathStr := rootName
	if len(path) > 0 {
		pathStr += " " + strings.Join(path, " ")
	}

	fields := helpFields{
		Indent:      indentStr,
		Path:        pathStr,
		ArgUsage:    usageText(cmd),
		Tagline:     cmd.Helptext.Tagline,
		Arguments:   cmd.Helptext.Arguments,
		Options:     cmd.Helptext.Options,
		Synopsis:    cmd.Helptext.Synopsis,
		Subcommands: cmd.Helptext.Subcommands,
		Description: cmd.Helptext.ShortDescription,
		Usage:       cmd.Helptext.Usage,
		MoreHelp:    (cmd != root),
	}

	if len(cmd.Helptext.LongDescription) > 0 {
		fields.Description = cmd.Helptext.LongDescription
	}

	// autogen fields that are empty
	if len(fields.Arguments) == 0 {
		fields.Arguments = strings.Join(argumentText(cmd), "\n")
	}
	if len(fields.Options) == 0 {
		fields.Options = strings.Join(optionText(cmd), "\n")
	}
	if len(fields.Subcommands) == 0 {
		fields.Subcommands = strings.Join(subcommandText(cmd, rootName, path), "\n")
	}

	// trim the extra newlines (see TrimNewlines doc)
	fields.TrimNewlines()

	// indent all fields that have been set
	fields.IndentAll()

	return longHelpTemplate.Execute(out, fields)
}

// ShortHelp returns a formatted CLI helptext string, generated for the given command
func ShortHelp(rootName string, root *cmds.Command, path []string, out io.Writer) error {
	cmd, err := root.Get(path)
	if err != nil {
		return err
	}

	// default cmd to root if there is no path
	if path == nil && cmd == nil {
		cmd = root
	}

	pathStr := rootName
	if len(path) > 0 {
		pathStr += " " + strings.Join(path, " ")
	}

	fields := helpFields{
		Indent:      indentStr,
		Path:        pathStr,
		ArgUsage:    usageText(cmd),
		Tagline:     cmd.Helptext.Tagline,
		Synopsis:    cmd.Helptext.Synopsis,
		Description: cmd.Helptext.ShortDescription,
		Usage:       cmd.Helptext.Usage,
		MoreHelp:    (cmd != root),
	}

	// trim the extra newlines (see TrimNewlines doc)
	fields.TrimNewlines()

	// indent all fields that have been set
	fields.IndentAll()

	return shortHelpTemplate.Execute(out, fields)
}

func argumentText(cmd *cmds.Command) []string {
	lines := make([]string, len(cmd.Arguments))

	for i, arg := range cmd.Arguments {
		lines[i] = argUsageText(arg)
	}
	lines = align(lines)
	for i, arg := range cmd.Arguments {
		lines[i] += " - " + arg.Description
	}

	return lines
}

func optionText(cmd ...*cmds.Command) []string {
	// get a slice of the options we want to list out
	options := make([]cmds.Option, 0)
	for _, c := range cmd {
		for _, opt := range c.Options {
			options = append(options, opt)
		}
	}

	// add option names to output (with each name aligned)
	lines := make([]string, 0)
	j := 0
	for {
		done := true
		i := 0
		for _, opt := range options {
			if len(lines) < i+1 {
				lines = append(lines, "")
			}

			names := sortByLength(opt.Names())
			if len(names) >= j+1 {
				lines[i] += fmt.Sprintf(optionFlag, names[j])
			}
			if len(names) > j+1 {
				lines[i] += ", "
				done = false
			}

			i++
		}

		if done {
			break
		}

		lines = align(lines)
		j++
	}
	lines = align(lines)

	// add option types to output
	for i, opt := range options {
		lines[i] += " " + fmt.Sprintf("%v", opt.Type())
	}
	lines = align(lines)

	// add option descriptions to output
	for i, opt := range options {
		lines[i] += " - " + opt.Description()
	}

	return lines
}

func subcommandText(cmd *cmds.Command, rootName string, path []string) []string {
	prefix := fmt.Sprintf("%v %v", rootName, strings.Join(path, " "))
	if len(path) > 0 {
		prefix += " "
	}
	subcmds := make([]*cmds.Command, len(cmd.Subcommands))
	lines := make([]string, len(cmd.Subcommands))

	i := 0
	for name, sub := range cmd.Subcommands {
		usage := usageText(sub)
		if len(usage) > 0 {
			usage = " " + usage
		}
		lines[i] = prefix + name + usage
		subcmds[i] = sub
		i++
	}

	lines = align(lines)
	for i, sub := range subcmds {
		lines[i] += " - " + sub.Helptext.Tagline
	}

	return lines
}

func usageText(cmd *cmds.Command) string {
	s := ""
	for i, arg := range cmd.Arguments {
		if i != 0 {
			s += " "
		}
		s += argUsageText(arg)
	}

	return s
}

func argUsageText(arg cmds.Argument) string {
	s := arg.Name

	if arg.Required {
		s = fmt.Sprintf(requiredArg, s)
	} else {
		s = fmt.Sprintf(optionalArg, s)
	}

	if arg.Variadic {
		s = fmt.Sprintf(variadicArg, s)
	}

	return s
}

func align(lines []string) []string {
	longest := 0
	for _, line := range lines {
		length := len(line)
		if length > longest {
			longest = length
		}
	}

	for i, line := range lines {
		length := len(line)
		if length > 0 {
			lines[i] += strings.Repeat(" ", longest-length)
		}
	}

	return lines
}

func indent(lines []string, prefix string) []string {
	for i, line := range lines {
		lines[i] = prefix + indentString(line, prefix)
	}
	return lines
}

func indentString(line string, prefix string) string {
	return prefix + strings.Replace(line, "\n", "\n"+prefix, -1)
}

type lengthSlice []string

func (ls lengthSlice) Len() int {
	return len(ls)
}
func (ls lengthSlice) Swap(a, b int) {
	ls[a], ls[b] = ls[b], ls[a]
}
func (ls lengthSlice) Less(a, b int) bool {
	return len(ls[a]) < len(ls[b])
}

func sortByLength(slice []string) []string {
	output := make(lengthSlice, len(slice))
	for i, val := range slice {
		output[i] = val
	}
	sort.Sort(output)
	return []string(output)
}
