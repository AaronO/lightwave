package lightwave

import (
  "regexp"
)

type UserId struct {
  Username string
  Domain string
}

func NewUserId(userid string) *UserId {
  r := regexp.MustCompile("^([\\-_A-Za-z0-9.]+)@([\\-_A-Za-z0-9.]+)$")
  if submatches := r.FindStringSubmatch(userid); submatches != nil {
	return &UserId{submatches[1], submatches[2]}
  }
  return nil
}