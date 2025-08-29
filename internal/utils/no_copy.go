package utils

import "sync"

// NoCopy prevents copying structs by accident. Adding it to a struct will cause go vet to flag it as an error when
// you try to copy the struct. This is inspired by sync.noCopy which is used in several concurrency primitives but
// not publicly exposed.
type NoCopy struct {}

// NoCopy implements sync.Locker
var _ sync.Locker = (*NoCopy)(nil)

func (n *NoCopy) Lock() {}

func (n *NoCopy) Unlock() {}
