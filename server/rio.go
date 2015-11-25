package server

type Rio interface {
    Write([]byte) (int, error)
    Read(length int) ([]byte, error)
    Tell() int64
    Flush() error
}

/*for rdb file*/
type RioFileIo struct {

}

/*for aof*/
type RioBufferIO struct {

}

/*for rdb slave*/
type RioSocksIo struct {

}
