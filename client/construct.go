package client

type Contructor interface{
    Construct(raw []byte) string
}

