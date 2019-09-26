# fsmtp

A simple SMTP server that will save all the send mail files content to a given location

## Installation
`go get -u -v github.com/pijalu/fsmtp`

## Usage
Usage:
  fsmtp [flags]

Flags:
  -a, --address string   Address to bind to (default "localhost")
      --config string    config file (default is $HOME/.fsmtp.yaml)
  -h, --help             help for fsmtp
  -o, --output string    Output location for the files (default ".")
  -p, --port int         Server port (default 2525)
  -t, --toggle           Help message for toggle

## Example
```sh 
$ fsmtp
2019/09/26 13:47:45 Configured Output path: .
2019/09/26 13:47:45 Starting up listening to localhost:2525
...
```

Send a mail using [smtp-cli](http://www.logix.cz/michal/devel/smtp-cli/)

```sh
$ smtp-cli --server localhost:2525 --verbose --from bla@bla.com --to bli@bla.com --subject "hoho" --body-plain "That a mail" -4 --attach=Documents/marzluff.png
```

this will result in the following line in the logs:
```
2019/09/26 14:23:02 Saved file ce/ce96c35e-ab62-4109-b59e-eab1ba05601a/marzluff.png [52911 byte(s)]
```

The file will be saved under 'output' (default: where the server is started), with a unique path. In this case 
```
ce/ce96c35e-ab62-4109-b59e-eab1ba05601a/marzluff.png
```

