package config

import (
	"testing"

	"github.com/matryer/is"
)

func TestParsesConfigPipeHCL(t *testing.T) {
	is := is.New(t)

	c, err := ParseHCLFile("../test_fixtures/pipe/standard.hcl")

	is.NoErr(err)             // error should have been nil
	is.Equal(1, len(c.Pipes)) // should have returned one pipe
}

func TestParsesConfigPipeHCLNoFail(t *testing.T) {
	is := is.New(t)

	c, err := ParseHCLFile("../test_fixtures/pipe/no_fail.hcl")

	is.NoErr(err)                                          // error should have been nil
	is.Equal(1, len(c.Pipes))                              // should have returned one pipe
	is.Equal(0, len(c.Pipes["process_image_fail"].OnFail)) // should have returned 0 fail blocks
}

func TestParsesConfigPipeHCLNoSuccess(t *testing.T) {
	is := is.New(t)

	c, err := ParseHCLFile("../test_fixtures/pipe/no_success.hcl")

	is.NoErr(err)                                                // error should have been nil
	is.Equal(1, len(c.Pipes))                                    // should have returned one pipe
	is.Equal(0, len(c.Pipes["process_image_success"].OnSuccess)) // should have returned 0 success blocks
}

func TestParsesFolder(t *testing.T) {
	is := is.New(t)

	c, err := ParseFolder("../test_fixtures")

	is.NoErr(err)               // error should have been nil
	is.Equal(3, len(c.Pipes))   // should have returned three pipes
	is.Equal(2, len(c.Outputs)) // should have returned two output
	is.Equal(1, len(c.Inputs))  // should have returned one input
}
