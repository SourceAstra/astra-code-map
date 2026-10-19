package core

import "testing"

func TestEstimateFunctionEndLine(t *testing.T) {
	cases := []struct {
		name  string
		lines []string
		start int
		want  int
	}{
		{"plain function", []string{
			"int f(void) {",
			"    return 0;",
			"}",
		}, 1, 3},
		{"braces in single-line string", []string{
			"int f(void) {",
			"    cmd(\"a {b} c\");",
			"    return 0;",
			"}",
		}, 1, 4},
		// Regression: CLI syntax literals continue across lines via trailing
		// backslashes; the string opened on line 2 carries over line 3, whose
		// "{static" must not count as a code brace (firefly_l2_cli.c).
		{"continued string braces ignored", []string{
			"FIREFLY_CLI(f, f_cmd,",
			"    \"l2 fdb add {port P | nexthop N} \\",
			"   ({static | discard}|)\",",
			"    help, help)",
			"{",
			"    call();",
			"}",
		}, 1, 7},
		{"escaped quote keeps string open", []string{
			"int f(void) {",
			"    s = \"a\\\"b}c\";",
			"}",
		}, 1, 3},
		{"line comment brace ignored", []string{
			"int f(void) { // {",
			"}",
		}, 1, 2},
		// Regression: block comments may span lines and carry braces; the
		// "}" before the closing */ must not count (firefly_l2_cli.c parse
		// section headers).
		{"block comment brace ignored", []string{
			"int f(void) {",
			"    /*4) parse static|discard|src-discard",
			"    |copy-tocpu|bind-port} */",
			"    return 0;",
			"}",
		}, 1, 5},
	}
	for _, tc := range cases {
		if got := estimateFunctionEndLine(tc.lines, tc.start); got != tc.want {
			t.Errorf("%s: got end line %d, want %d", tc.name, got, tc.want)
		}
	}
}
