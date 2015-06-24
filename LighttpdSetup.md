# Lighttpd Configuration #

The following setup allows for testing wave federation on a single computer.
Setup lighttpd with FCGI support.

Add the following to `/etc/hosts/`:

```
127.0.0.1       w1.com
127.0.0.1       w2.com
```

Using URL rewriting we direct wave requests to two different lightwave instances:

```

$HTTP["host"] == "w2.com" {
  url.rewrite                 = ( "^/wave/(.+)$" => "/vave/$1" )
}

server.modules   += ( "mod_fastcgi" )

fastcgi.server  = ( "/wave/" =>
    ( "localhost" =>
      ( "host" => "127.0.0.1"
      , "port" => 9871
      , "check-local" => "disable"
      , "min-procs" => 1
      , "max-procs" => 1
      )
    ),
        "/vave/" =>
    ( "w2.com" =>
      ( "host" => "127.0.0.1"
      , "port" => 9872
      , "check-local" => "disable"
      , "min-procs" => 1
      , "max-procs" => 1
      )
    )    
  )
```

Now you can start two instances of lightwave as follows:
```
lighttpd -p 9871 -d w1.com
lighttpd -p 9872 -d w2.com
```