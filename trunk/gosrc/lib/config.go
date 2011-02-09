package lightwave

import (
  "os"
  "json"
)

type Config struct {
  /**
   * Relative paths in the file system which point to server configuration files.
   */
  Servers []string
  /**
   * Port on which to listen for incoming HTTP connections.
   */
  Port uint16
}

type ServerConfig struct {
  MainConfig *Config
  Domain string
  Hostname string
  /**
   * Name of a directory (not ending with '/') that stores all documents, i.e.
   * snapshot, deltas, etc.
   */
  DataRoot string
  /**
   * Name of a directory (not ending with '/') that stores static files, e.g.
   * the index.html page, images etc. These files are available without authentication
   */
  StaticRoot string
  /**
   * The relative URL (no host, no transport) to which the user should be redirected after logging in,
   * e.g. "/index.html".
   */
  MainPage string
  /**
   * The relative URL (no host, no transport) for the login page,
   * e.g. "/login.html".
   */
  LoginPage string
  /**
   * The relative URL (no host, no transport) for the registration page,
   * e.g. "/signup.html".
   */
  SignupPage string
  /**
   * Filename to the sqlite database that stores user names.
   */
  UserDB string
  /**
   * Filename to the sqlite database that stores the index.
   */
  IndexDB string
}

func ReadConfig() (*Config, os.Error) {
  file, err := os.Open("lightwave.config", os.O_RDONLY, 0700);
  if err != nil {
    return nil, err
  }
  stat, err := file.Stat();
  if err != nil {
    return nil, err
  }
  bytes := make([]byte, stat.Size)
  n, err := file.Read(bytes)
  if err != nil || n != len(bytes) {
    return nil, err
  }
  config := &Config{}
  err = json.Unmarshal(bytes, config)
  if err != nil {
    return nil, err
  }
  return config, nil
}

func ReadServerConfig(config *Config, filename string) (*ServerConfig, os.Error) {
  file, err := os.Open(filename, os.O_RDONLY, 0700);
  if err != nil {
    return nil, err
  }
  stat, err := file.Stat();
  if err != nil {
    return nil, err
  }
  bytes := make([]byte, stat.Size)
  n, err := file.Read(bytes)
  if err != nil || n != len(bytes) {
    return nil, err
  }
  serverconfig := &ServerConfig{}
  err = json.Unmarshal(bytes, serverconfig)
  if err != nil {
    return nil, err
  }
  serverconfig.MainConfig = config
  return serverconfig, nil
}