package lightwave

import (
  "strings"
)

// -----------------------------------------------
// URI

type URI interface {
  String() string
}

// -----------------------------------------------
// ViewURI

type ViewURI struct {
  Name string
}

func (self ViewURI) String() string {
  return "_view/" + self.Name
}

// -----------------------------------------------
// ManifestURI

type ManifestURI struct {
}

func (self ManifestURI) String() string {
  return "_manifest"
}

// -----------------------------------------------
// DocumentURI

type DocumentURI struct {
  Host string
  NameSeq []string
}

func (self DocumentURI) String() string {
  return self.Host + "/" + strings.Join(self.NameSeq, "/")
}

// -----------------------------------------------
// StaticURI

type StaticURI struct {
  NameSeq []string
}

func (self StaticURI) String() string {
  return "/_static/" + strings.Join(self.NameSeq, "/")
}

// -----------------------------------------------
// Constructor

func NewURI(uri string) (result URI, ok bool) {
  if uri[0] != '/' {
	return nil, false
  }
  uri = uri[1:]
  if uri == "_manifest" {
	return ManifestURI{}, true
  }
  // Strip a trailing slash
  if strings.HasSuffix(uri, "/") {
	uri = uri[0:len(uri)-1]
  }
  slices := strings.Split( uri, "/", -1 )
  for _, s := range slices {
	if s == "" {
	  return nil, false
	}
  }
  // TODO: Ensure that no ".." sequences are used
  if slices[0] == "_static" {
	return StaticURI{NameSeq:slices[1:]}, true
  }
  if len(slices) == 0 {
	return nil, false
  }
  if slices[0] == "_view" {
	if len(slices) != 2 {
	  return nil, false
	}
	return ViewURI{slices[1]}, true
  } else if slices[0][0] != '_' {
	if len(slices) < 2 {
	  return nil, false
	}
	return DocumentURI{slices[0], slices[1:]}, true
  }

  return nil, false
}
