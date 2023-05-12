package crypto_service

import "fmt"

type dummyCapability struct {
	id              string
	name            string
	description     string
	numArguments    uint32
	shouldSucceed   bool
	result          [][]byte
	errorMessage    string
	computeCallsLog []logEntry
}

type logEntry struct {
	in [][]byte
}

func newDummyCapability(
	id string,
	name string,
	description string,
	numArguments uint32,
) *dummyCapability {
	return &dummyCapability{
		id:              id,
		name:            name,
		description:     description,
		numArguments:    numArguments,
		shouldSucceed:   true,
		result:          [][]byte{[]byte("Mickey"), {}, []byte("Mouse")},
		errorMessage:    "default error message",
		computeCallsLog: []logEntry{},
	}
}

func (c *dummyCapability) Id() string {
	return c.id
}

func (c *dummyCapability) Name() string {
	return c.name
}

func (c *dummyCapability) Description() string {
	return c.description
}

func (c *dummyCapability) NumArguments() uint32 {
	return c.numArguments
}

func (c *dummyCapability) Compute(in [][]byte) ([][]byte, error) {
	entry := logEntry{in: in}
	c.computeCallsLog = append(c.computeCallsLog, entry)
	if c.shouldSucceed {
		return c.getResult(), nil
	} else {
		return c.getResult(), fmt.Errorf(c.getErrorMessage())
	}
}

func (c *dummyCapability) getNumCalls() int {
	return len(c.computeCallsLog)
}

func (c *dummyCapability) getResult() [][]byte {
	return c.result
}

func (c *dummyCapability) getErrorMessage() string {
	return c.errorMessage
}
