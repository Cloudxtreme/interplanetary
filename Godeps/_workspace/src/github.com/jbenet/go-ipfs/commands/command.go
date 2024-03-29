package commands

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	u "github.com/maybebtc/interplanetary/Godeps/_workspace/src/github.com/jbenet/go-ipfs/util"
)

var log = u.Logger("command")

// Function is the type of function that Commands use.
// It reads from the Request, and writes results to the Response.
type Function func(Request) (interface{}, error)

// Marshaler is a function that takes in a Response, and returns a marshalled []byte
// (or an error on failure)
type Marshaler func(Response) ([]byte, error)

// MarshalerMap is a map of Marshaler functions, keyed by EncodingType
// (or an error on failure)
type MarshalerMap map[EncodingType]Marshaler

// HelpText is a set of strings used to generate command help text. The help
// text follows formats similar to man pages, but not exactly the same.
type HelpText struct {
	// required
	Tagline          string // used in <cmd usage>
	ShortDescription string // used in DESCRIPTION
	Synopsis         string // showcasing the cmd

	// optional - whole section overrides
	Usage           string // overrides USAGE section
	LongDescription string // overrides DESCRIPTION section
	Options         string // overrides OPTIONS section
	Arguments       string // overrides ARGUMENTS section
	Subcommands     string // overrides SUBCOMMANDS section
}

// TODO: check Argument definitions when creating a Command
//   (might need to use a Command constructor)
//   * make sure any variadic args are at the end
//   * make sure there aren't duplicate names
//   * make sure optional arguments aren't followed by required arguments

// Command is a runnable command, with input arguments and options (flags).
// It can also have Subcommands, to group units of work into sets.
type Command struct {
	Options    []Option
	Arguments  []Argument
	Run        Function
	Marshalers map[EncodingType]Marshaler
	Helptext   HelpText

	// Type describes the type of the output of the Command's Run Function.
	// In precise terms, the value of Type is an instance of the return type of
	// the Run Function.
	//
	// ie. If command Run returns &Block{}, then Command.Type == &Block{}
	Type        interface{}
	Subcommands map[string]*Command
}

// ErrNotCallable signals a command that cannot be called.
var ErrNotCallable = ClientError("This command can't be called directly. Try one of its subcommands.")

var ErrNoFormatter = ClientError("This command cannot be formatted to plain text")

var ErrIncorrectType = errors.New("The command returned a value with a different type than expected")

// Call invokes the command for the given Request
func (c *Command) Call(req Request) Response {
	res := NewResponse(req)

	cmds, err := c.Resolve(req.Path())
	if err != nil {
		res.SetError(err, ErrClient)
		return res
	}
	cmd := cmds[len(cmds)-1]

	if cmd.Run == nil {
		res.SetError(ErrNotCallable, ErrClient)
		return res
	}

	err = cmd.CheckArguments(req)
	if err != nil {
		res.SetError(err, ErrClient)
		return res
	}

	err = req.ConvertOptions()
	if err != nil {
		res.SetError(err, ErrClient)
		return res
	}

	output, err := cmd.Run(req)
	if err != nil {
		// if returned error is a commands.Error, use its error code
		// otherwise, just default the code to ErrNormal
		switch e := err.(type) {
		case *Error:
			res.SetError(e, e.Code)
		case Error:
			res.SetError(e, e.Code)
		default:
			res.SetError(err, ErrNormal)
		}
		return res
	}

	// If the command specified an output type, ensure the actual value returned is of that type
	if cmd.Type != nil {
		definedType := reflect.ValueOf(cmd.Type).Type()
		actualType := reflect.ValueOf(output).Type()

		if definedType != actualType {
			res.SetError(ErrIncorrectType, ErrNormal)
			return res
		}
	}

	// clean up the request (close the readers, e.g. fileargs)
	// NOTE: this means commands can't expect to keep reading after cmd.Run returns (in a goroutine)
	err = req.Cleanup()
	if err != nil {
		res.SetError(err, ErrNormal)
		return res
	}

	res.SetOutput(output)
	return res
}

// Resolve gets the subcommands at the given path
func (c *Command) Resolve(path []string) ([]*Command, error) {
	cmds := make([]*Command, len(path)+1)
	cmds[0] = c

	cmd := c
	for i, name := range path {
		cmd = cmd.Subcommand(name)

		if cmd == nil {
			pathS := strings.Join(path[0:i], "/")
			return nil, fmt.Errorf("Undefined command: '%s'", pathS)
		}

		cmds[i+1] = cmd
	}

	return cmds, nil
}

// Get resolves and returns the Command addressed by path
func (c *Command) Get(path []string) (*Command, error) {
	cmds, err := c.Resolve(path)
	if err != nil {
		return nil, err
	}
	return cmds[len(cmds)-1], nil
}

// GetOptions gets the options in the given path of commands
func (c *Command) GetOptions(path []string) (map[string]Option, error) {
	options := make([]Option, 0, len(c.Options))

	cmds, err := c.Resolve(path)
	if err != nil {
		return nil, err
	}
	cmds = append(cmds, globalCommand)

	for _, cmd := range cmds {
		options = append(options, cmd.Options...)
	}

	optionsMap := make(map[string]Option)
	for _, opt := range options {
		for _, name := range opt.Names() {
			if _, found := optionsMap[name]; found {
				return nil, fmt.Errorf("Option name '%s' used multiple times", name)
			}

			optionsMap[name] = opt
		}
	}

	return optionsMap, nil
}

func (c *Command) CheckArguments(req Request) error {
	args := req.Arguments()

	// count required argument definitions
	numRequired := 0
	for _, argDef := range c.Arguments {
		if argDef.Required {
			numRequired++
		}
	}

	// iterate over the arg definitions
	valueIndex := 0 // the index of the current value (in `args`)
	for _, argDef := range c.Arguments {
		// skip optional argument definitions if there aren't sufficient remaining values
		if len(args)-valueIndex <= numRequired && !argDef.Required || argDef.Type == ArgFile {
			continue
		}

		// the value for this argument definition. can be nil if it wasn't provided by the caller
		v, found := "", false
		if valueIndex < len(args) {
			v = args[valueIndex]
			found = true
			valueIndex++
		}

		err := checkArgValue(v, found, argDef)
		if err != nil {
			return err
		}

		// any additional values are for the variadic arg definition
		if argDef.Variadic && valueIndex < len(args)-1 {
			for _, val := range args[valueIndex:] {
				err := checkArgValue(val, true, argDef)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// Subcommand returns the subcommand with the given id
func (c *Command) Subcommand(id string) *Command {
	return c.Subcommands[id]
}

// checkArgValue returns an error if a given arg value is not valid for the given Argument
func checkArgValue(v string, found bool, def Argument) error {
	if !found && def.Required {
		return fmt.Errorf("Argument '%s' is required", def.Name)
	}

	return nil
}

func ClientError(msg string) error {
	return &Error{Code: ErrClient, Message: msg}
}
