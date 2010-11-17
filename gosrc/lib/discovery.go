package lightwave

import (
  "http"
  "net"
  "log"
  "fmt"
)

//---------------------------------------------------
// Discovery

func GetManifest(host string, port uint16) *ServerManifest {  
  log.Println("Discovery at ", host, port)
  // Make a TCP connection
  con, err := net.Dial("tcp", "", fmt.Sprintf("%v:%v", host, port))
  if err != nil {
	log.Println("Failed connecting to ", host, port, err)
	return nil
  }
  // Make a HTTP connection
  hcon := http.NewClientConn(con, nil)
  // Build the HTTP request
  var hreq http.Request
  hreq.Host = host
  hreq.Method = "GET"  
  hreq.RawURL = fmt.Sprintf("http://%v:%v/_manifest", host, port)
  log.Println("Sending request")  
  // Send the HTTP request
  if err = hcon.Write(&hreq); err != nil {
	log.Println("Error sending GET request")
	hcon.Close()
	return nil
  }
  // Read the HTTP response
  _, err = hcon.Read()
  if err != nil {
	log.Println("Error reading HTTP response from ", host, err)
	hcon.Close()
	return nil
  }
  
  m := &ServerManifest{}
  // TODO: Parse the manifest
  // HACK START
  m.HostName = host
  m.Port = port
  // HACK END
  return m
}

func Discover(domain string) *ServerManifest {
  for {
	// Step 1: DNS SRV lookup
	cname, addrs, err := net.LookupSRV("wave-server", "tcp", domain)
	if err == nil && len(addrs) > 0 {
	  log.Println("DNS SRV lookup: ", cname, addrs[0].Target, addrs[0].Port)
	  if manifest := GetManifest(addrs[0].Target, addrs[0].Port); manifest != nil {
		// HACK START
		manifest.Domain = domain
		// HACK END
		return manifest
	  }
	}
	// Step 2: Lookup of wave.domain.com:80
	if manifest := GetManifest("wave." + domain, 8080); manifest != nil {
	  // HACK START
	  manifest.Domain = domain
	  // HACK END
	  return manifest
	}
	// Step 3: Lookup domain.com:80
	if manifest := GetManifest(domain, 8080); manifest != nil {
	  // HACK START
	  manifest.Domain = domain
	  // HACK END
	  return manifest
	}
	// Discovery failed
	return nil
  }
  return nil
}
