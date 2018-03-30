package main

// 	(c) Holger Berger 2018

import (
	"regexp"
	"strings"
)

// Opstable contains all ops
type Opstable struct {
	ops map[string]string
	re  *regexp.Regexp
}

// NewOpstable creates table expanding x.y.z to x.y and x.z, cutting at first [ or space
func NewOpstable() *Opstable {
	var no Opstable

	no.ops = make(map[string]string)

	// FIXME what avout , in first position?
	no.re = regexp.MustCompile(`^(.+?)[\[\s].*$`)

	// FIXME hard coded veops
	for _, o := range veops {
		m := no.re.FindStringSubmatch(o)
		if m != nil {
			t := strings.Split(m[1], ".")
			if len(t) > 1 {
				for i := 1; i < len(t); i++ {
					no.ops[t[0]+"."+t[i]] = o
				}
				no.ops[t[0]] = o
			} else {
				no.ops[t[0]] = o
			}
		}
	}

	// FIXME hardcoded VE
	for co := range bcodes {
		op := "b" + co
		no.ops[op] = "= Branch on condition " + bcodes[co]
		op = "br" + co
		no.ops[op] = "= Branch relative on condition " + bcodes[co]
	}

	return &no
}

// getops finds ops in opstable
func (o *Opstable) getops(ops string) string {
	e, ok := o.ops[ops]
	if ok {
		return e
	}
	return ""
}
