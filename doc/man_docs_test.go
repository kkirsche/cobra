package doc

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

var _ = fmt.Println
var _ = os.Stderr

func translate(in string) string {
	return strings.Replace(in, "-", "\\-", -1)
}

func TestGenManDoc(t *testing.T) {
	c := initializeWithRootCmd()
	// Need two commands to run the command alphabetical sort
	cmdEcho.AddCommand(cmdTimes, cmdEchoSub, cmdDeprecated)
	c.AddCommand(cmdPrint, cmdEcho)
	cmdRootWithRun.PersistentFlags().StringVarP(&flags2a, "rootflag", "r", "two", strtwoParentHelp)

	out := new(bytes.Buffer)

	header := &GenManHeader{
		Title:   "Project",
		Section: "2",
	}
	// We generate on a subcommand so we have both subcommands and parents
	if err := GenMan(cmdEcho, header, out); err != nil {
		t.Fatal(err)
	}
	found := out.String()

	// Make sure parent has - in CommandPath() in SEE ALSO:
	parentPath := cmdEcho.Parent().CommandPath()
	dashParentPath := strings.Replace(parentPath, " ", "-", -1)
	expected := translate(dashParentPath)
	expected = expected + "(" + header.Section + ")"
	checkStringContains(t, found, expected)

	// Our description
	expected = translate(cmdEcho.Name())
	checkStringContains(t, found, expected)

	// Better have our example
	expected = translate(cmdEcho.Name())
	checkStringContains(t, found, expected)

	// A local flag
	expected = "boolone"
	checkStringContains(t, found, expected)

	// persistent flag on parent
	expected = "rootflag"
	checkStringContains(t, found, expected)

	// We better output info about our parent
	expected = translate(cmdRootWithRun.Name())
	checkStringContains(t, found, expected)

	// And about subcommands
	expected = translate(cmdEchoSub.Name())
	checkStringContains(t, found, expected)

	unexpected := translate(cmdDeprecated.Name())
	checkStringOmits(t, found, unexpected)

	// auto generated
	expected = translate("Auto generated")
	checkStringContains(t, found, expected)
}

func TestGenManNoGenTag(t *testing.T) {
	c := initializeWithRootCmd()
	// Need two commands to run the command alphabetical sort
	cmdEcho.AddCommand(cmdTimes, cmdEchoSub, cmdDeprecated)
	c.AddCommand(cmdPrint, cmdEcho)
	cmdRootWithRun.PersistentFlags().StringVarP(&flags2a, "rootflag", "r", "two", strtwoParentHelp)
	cmdEcho.DisableAutoGenTag = true
	out := new(bytes.Buffer)

	header := &GenManHeader{
		Title:   "Project",
		Section: "2",
	}
	// We generate on a subcommand so we have both subcommands and parents
	if err := GenMan(cmdEcho, header, out); err != nil {
		t.Fatal(err)
	}
	found := out.String()

	unexpected := translate("#HISTORY")
	checkStringOmits(t, found, unexpected)
}

func TestGenManSeeAlso(t *testing.T) {
	noop := func(cmd *cobra.Command, args []string) {}

	top := &cobra.Command{Use: "top", Run: noop}
	aaa := &cobra.Command{Use: "aaa", Run: noop, Hidden: true} // #229
	bbb := &cobra.Command{Use: "bbb", Run: noop}
	ccc := &cobra.Command{Use: "ccc", Run: noop}
	top.AddCommand(aaa, bbb, ccc)

	out := new(bytes.Buffer)
	header := &GenManHeader{}
	if err := GenMan(top, header, out); err != nil {
		t.Fatal(err)
	}

	scanner := bufio.NewScanner(out)

	if err := AssertLineFound(scanner, ".SH SEE ALSO"); err != nil {
		t.Fatal(fmt.Errorf("Couldn't find SEE ALSO section header: %s", err.Error()))
	}

	if err := AssertNextLineEquals(scanner, ".PP"); err != nil {
		t.Fatal(fmt.Errorf("First line after SEE ALSO wasn't break-indent: %s", err.Error()))
	}

	if err := AssertNextLineEquals(scanner, `\fBtop\-bbb(1)\fP, \fBtop\-ccc(1)\fP`); err != nil {
		t.Fatal(fmt.Errorf("Second line after SEE ALSO wasn't correct: %s", err.Error()))
	}
}

func AssertLineFound(scanner *bufio.Scanner, expectedLine string) error {
	for scanner.Scan() {
		line := scanner.Text()
		if line == expectedLine {
			return nil
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("AssertLineFound: scan failed: %s", err.Error())
	}

	return fmt.Errorf("AssertLineFound: hit EOF before finding %#v", expectedLine)
}

func AssertNextLineEquals(scanner *bufio.Scanner, expectedLine string) error {
	if scanner.Scan() {
		line := scanner.Text()
		if line == expectedLine {
			return nil
		} else {
			return fmt.Errorf("AssertNextLineEquals: got %#v, not %#v", line, expectedLine)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("AssertNextLineEquals: scan failed: %s", err.Error())
	}

	return fmt.Errorf("AssertNextLineEquals: hit EOF before finding %#v", expectedLine)
}
