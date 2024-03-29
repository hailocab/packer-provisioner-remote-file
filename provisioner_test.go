package main

import (
	"github.com/mitchellh/packer/packer"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func testConfig() map[string]interface{} {
	return map[string]interface{}{
		"destination": "something",
	}
}

func TestProvisioner_Impl(t *testing.T) {
	var raw interface{}
	raw = &Provisioner{}
	if _, ok := raw.(packer.Provisioner); !ok {
		t.Fatalf("must be a provisioner")
	}
}

func TestProvisionerPrepare_InvalidKey(t *testing.T) {
	var p Provisioner
	config := testConfig()

	// Add a random key
	config["i_should_not_be_valid"] = true
	err := p.Prepare(config)
	if err == nil {
		t.Fatal("should have error")
	}
}

func TestProvisionerPrepare_ValidSource(t *testing.T) {
	var p Provisioner

	tf, err := ioutil.TempFile("", "packer")
	if err != nil {
		t.Fatalf("error tempfile: %s", err)
	}
	defer os.Remove(tf.Name())

	config := testConfig()
	config["source"] = tf.Name()

	err = p.Prepare(config)
	if err != nil {
		t.Fatalf("should allow valid file: %s", err)
	}
}

func TestProvisionerPrepare_EmptyDestination(t *testing.T) {
	var p Provisioner

	config := testConfig()
	delete(config, "destination")
	err := p.Prepare(config)
	if err == nil {
		t.Fatalf("should require destination path")
	}
}

type stubUi struct {
	sayMessages string
}

func (su *stubUi) Ask(string) (string, error) {
	return "", nil
}

func (su *stubUi) Error(string) {
}

func (su *stubUi) Machine(string, ...string) {
}

func (su *stubUi) Message(string) {
}

func (su *stubUi) Say(msg string) {
	su.sayMessages += msg
}

func TestProvisionerProvision_SendsFile(t *testing.T) {
	var p Provisioner
	tf, err := ioutil.TempFile("", "packer")
	if err != nil {
		t.Fatalf("error tempfile: %s", err)
	}
	defer os.Remove(tf.Name())

	if _, err = tf.Write([]byte("hello")); err != nil {
		t.Fatalf("error writing tempfile: %s", err)
	}

	config := map[string]interface{}{
		"source":      tf.Name(),
		"destination": "something",
	}

	if err := p.Prepare(config); err != nil {
		t.Fatalf("err: %s", err)
	}

	ui := &stubUi{}
	comm := &packer.MockCommunicator{}
	err = p.Provision(ui, comm)
	if err != nil {
		t.Fatalf("should successfully provision: %s", err)
	}

	if !strings.Contains(ui.sayMessages, tf.Name()) {
		t.Fatalf("should print source filename")
	}

	if !strings.Contains(ui.sayMessages, "something") {
		t.Fatalf("should print destination filename")
	}

	if comm.UploadPath != "something" {
		t.Fatalf("should upload to configured destination")
	}

	if comm.UploadData != "hello" {
		t.Fatalf("should upload with source file's data")
	}
}
