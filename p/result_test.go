package p

import (
	"bytes"
	"io/ioutil"
	"os/exec"
	"strings"
	"testing"
)

func TestResult(t *testing.T) {
	t.Helper()

	for _, s := range []string{
		// Sandbox
		`<?php echo "a";`,
		// examples/30.php
		`<?php

		$a = ['a', 'b', 'c'];

		foreach ($a as $k => $v) {
			echo "$k: $v\n";
		}

		foreach ([0, 1, 2, 3] as $k => $v) {
			echo "$k: $v\n";
		}

		foreach ([0, 1, 2, 3] as $v) {
			echo ": $v\n";
		}
		`,
		// examples/31.php
		`<?php

		$array = ['placeholder'];
		$array['e'] = '5';

		if (isset($array['e'])) {
			echo "\"e\": $array[e]\n";
		} else {
			echo "\"e\" is not set.\n";
		}

		unset($array['e']);

		if (isset($array['e'])) {
			echo "\"e\": $array[e]\n";
		} else {
			echo "\"e\" is not set.\n";
		}

		$array['e'] = 'f';
		if (isset($array['e'])) {
			echo "\"e\": $array[e]\n";
		} else {
			echo "\"e\" is not set.\n";
		}
		`,
	} {
		src, err := ioutil.TempFile("", "php")
		if err != nil {
			t.Fatalf("opening temp file: %v", err)
		}
		defer src.Close()

		src.WriteString(s)

		cmd := exec.Command("php", "-f", src.Name())
		var refOut bytes.Buffer
		cmd.Stdout = &refOut
		err = cmd.Run()
		if err != nil {
			t.Fatalf("executing php: %v", err)
		}

		root, err := ioutil.TempDir("", "go-build")
		if err != nil {
			t.Fatalf("creating temp dir: %v", err)
		}

		ref, err := ioutil.TempFile(root, "")
		if err != nil {
			t.Fatalf("opening temp file: %v", err)
		}
		defer ref.Close()

		p := NewParser(NewNameTranslator(), NewFunctionTranslator())
		res := p.Run(parsePHP([]byte(s)))
		ref.WriteString(res.String())

		fileName := strings.TrimPrefix(ref.Name(), root+"/")
		cmd = exec.Command("mv", fileName, "main.go")
		cmd.Dir = root
		err = cmd.Run()
		if err != nil {
			t.Fatalf("changing directory: %v", err)
		}

		cmd = exec.Command("go", "run", "main.go")
		cmd.Dir = root

		var out bytes.Buffer
		var errf bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &errf
		err = cmd.Run()
		if err != nil {
			t.Error(errf.String())
			t.Fatalf("executing go: %v", err)
		}

		if refOut.String() != out.String() {
			t.Errorf("Expected:\n%s\nGot:\n%s\n", refOut.String(), out.String())
		}
	}
}
