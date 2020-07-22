package mots

type Author byte

const (
	AuthorClient Author = iota
	AuthorServer
)

func (author Author) String() string {
	var names = map[Author]string{
		AuthorClient: "Client",
		AuthorServer: "Server",
	}

	return names[author]
}

func (author Author) IsClient() bool {
	return author == AuthorClient
}

func (author Author) IsNotClient() bool {
	return author != AuthorClient
}

func (author Author) IsServer() bool {
	return author == AuthorServer
}

func (author Author) IsNotServer() bool {
	return author != AuthorServer
}
