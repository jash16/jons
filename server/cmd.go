package server

type cmdFunc func(cli *Client) error

type dirtyCmd struct {
    selectDb int32
    argv [][]byte
    argc int32
}
