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
// SessionURI

type SessionURI struct {
  User string
  Name string
  // The optional third path element which must start with '_'
  Special string
}

func (self SessionURI) String() string {
  if self.Special == "" {
	return "_session/" + self.User + "/" + self.Name
  }
  return "_session/" + self.User + "/" + self.Name + "/" + self.Special
}

// -----------------------------------------------
// UserURI

type UserURI struct {
  // A string of the form "joe", i.e. with no domain suffix
  User string
}

func (self UserURI) String() string {
  return "_user/" + self.User
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
  } else if slices[0] == "_user" {
	if len(slices) != 2 {
	  return nil, false
	}
	return UserURI{slices[1]}, true
  } else if slices[0] == "_session" {
	if len(slices) != 3 && len(slices) != 4 {
	  return nil, false
	}
	if len(slices) == 4 {
	  if (len(slices[3]) == 0 || slices[3][0] != '_') {
		return nil, false
	  }
	  return SessionURI{slices[1], slices[2], slices[3]}, true
	}
	return SessionURI{slices[1], slices[2], ""}, true
  } else if slices[0][0] != '_' {
	if len(slices) < 2 {
	  return nil, false
	}
	return DocumentURI{slices[0], slices[1:]}, true
  }

  return nil, false
}
