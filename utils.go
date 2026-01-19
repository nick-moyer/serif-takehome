package main

import "io"

type ByteCounter struct {
    io.Reader
    Count int64
}

func (bc *ByteCounter) Read(p []byte) (int, error) {
    n, err := bc.Reader.Read(p)
    bc.Count += int64(n)
    return n, err
}