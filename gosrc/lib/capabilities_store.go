package lightwave

type CapabilitiesStore struct {
  // Maps the domain name to the corresponding server manifest
  capabilities map[string]*ServerManifest
}

func NewCapabilitiesStore() *CapabilitiesStore {
  c := &CapabilitiesStore{}
  c.capabilities = make(map[string]*ServerManifest)
  return c
}

func (self *CapabilitiesStore) Find(domain string) *ServerManifest {
  c, ok := self.capabilities[domain]
  if !ok {
	return nil
  }
  return c
}

func (self *CapabilitiesStore) Insert(domain string, capabilities *ServerManifest) {
  self.capabilities[domain] = capabilities
}
