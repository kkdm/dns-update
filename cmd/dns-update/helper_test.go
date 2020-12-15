package main

import (
    "testing"

    "github.com/stretchr/testify/assert"
)

func TestValidateMsg(t *testing.T) {
    ok, err := validateMsg("example.com:127.0.0.1")
    assert.Equal(t, ok, true)
    assert.NoError(t, err)

    ng1, err := validateMsg("")
    assert.Equal(t, ng1, false)
    assert.Error(t, err)

    ng2, err := validateMsg("example.com:1277.0.0.1")
    assert.Equal(t, ng2, false)
    assert.Error(t, err)
}

func TestgetDomain(t *testing.T) {
    domainOK, err := getDomain("example.com:127.0.0.1")
    assert.Equal(t, domainOK, Domain{Name:"example.com", IP:"127.0.0.1"})
    assert.NoError(t, err)

    domainNG1, err := getDomain("")
    assert.Equal(t, domainNG1, Domain{})
    assert.Error(t, err)

    domainNG2, err := validateMsg("example.com:1277.0.0.1")
    assert.Equal(t, domainNG2, Domain{})
    assert.Error(t, err)
}
