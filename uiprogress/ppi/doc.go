// Package ppi contains the progress programming interface.
// It provides some standard implementations for progress elements
// and the used support type for new implementations.
//
// The support types use interface methods not intended
// to be used by the end user. To be usable by the support types
// the progress implementation must offer those methods as public
// methods. Therefore, the progress implementations are in a sub package
// while the official interface is described by interfaces offered
// in the main package for public use.
package ppi
