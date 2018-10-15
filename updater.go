package main

import (
  "github.com/mathewmoon/cfgo"
  "os"
  "fmt"
  "io/ioutil"
  "time"
  "log"
  "net/http"
  toml "github.com/BurntSushi/toml"
)

type Config struct {
  General GeneralOptions
  Cloudflare CloudflareOptions
}

type CloudflareOptions struct {
  Key string
  Email string
  Zone string
  Record string
  Ttl string
  RecordId string
  ZoneId string
  LogLocation string
  RecordType string
}

type GeneralOptions struct {
  Loglocation string
  Logname string
}

/*Start by getting all of the info we need for the first call */
func LoadConfig() (*Config, error) {
  var conf Config
  _, err := toml.DecodeFile("/etc/cf_dynamic_updater/updater.conf", &conf)
  if err != nil{
    fmt.Println(err)
    return nil, err
  }
  if conf.Cloudflare.Key == "" {
    fmt.Println("Option 'key' must be set in config")
    os.Exit(1)
  }
  if conf.Cloudflare.Zone == "" {
    fmt.Println("Option 'zone' must be set in config")
    os.Exit(1)
  }
  if conf.Cloudflare.Record == "" {
    fmt.Println("Option 'record' must be set in config")
    os.Exit(1)
  }
  if conf.Cloudflare.Email == "" {
    fmt.Println("Option 'email' must be set in config")
    os.Exit(1)
  }
  if conf.Cloudflare.RecordType == "" {
    fmt.Println("Option 'recordtype' must be set in config")
    os.Exit(1)
  }

  return &conf, nil
}

func getCurrentIp() (string, error) {
  resp, err := http.Get("https://api.ipify.org")
  if err != nil {
    return " ", err
  }else{
    body, err := ioutil.ReadAll(resp.Body)
    defer resp.Body.Close()
    if err != nil {
      return "", err
    }else{
      return string(body), nil
    }
  }
}

func main() {
  config,err := LoadConfig()
  if err != nil {
    fmt.Println(err)
    os.Exit(1)
  }

  /* Configure the logger */
  if _, err := os.Stat(config.General.Loglocation); os.IsNotExist(err) {
    fmt.Print("\n Log location does not exist.\n")
    os.Exit(1)
  }

  if _, err := os.Stat(config.General.Loglocation + "/" + config.General.Logname); os.IsNotExist(err) {
    //create your file with desired read/write permissions
    f, err := os.OpenFile(config.General.Loglocation + "/" + config.General.Logname, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
    if err != nil {
      fmt.Println(err)
      os.Exit(1)
    }
    //defer to close when you're done with it, not because you think it's idiomatic!
    defer f.Close()
    //set output of logs to f
    log.SetOutput(f)

  }else{
    f, err := os.OpenFile(config.General.Loglocation + "/" + config.General.Logname, os.O_WRONLY|os.O_APPEND, 0644)
    if err != nil {
      fmt.Println(err)
      os.Exit(1)
    }
    //defer to close when you're done with it, not because you think it's idiomatic!
    defer f.Close()
    //set output of logs to f
    log.SetOutput(f)
    log.Print("Starting Cloudflare dynamic DNS updater.")
  }


  /* Configure the client */
  var client = cfgo.Client{}
  client.Email = config.Cloudflare.Email
  client.Domain = config.Cloudflare.Zone
  client.Token = config.Cloudflare.Key

  /* Get the zone ID for future calls */
  domain, err := client.GetZone();
  if err != nil {
    log.Print(client.GetError())
    os.Exit(1)
  }

  config.Cloudflare.ZoneId = domain[0].ID

  /* Get the record ID for future calls */
  record, err := client.GetRecord(config.Cloudflare.Record, config.Cloudflare.RecordType)
  if err != nil {
    log.Print(err)
    os.Exit(1)
  }
  config.Cloudflare.RecordId = record[0].ID

  for 1 == 1 {
    record,err := client.GetSingleRecord(config.Cloudflare.ZoneId, config.Cloudflare.RecordId)
    if err != nil {
      continue
    }
    cfIp := record.Content

    currentIp,err := getCurrentIp()
    if err != nil {
      continue
    }

    if currentIp != cfIp {
      var data = []byte(`{"type": "` + config.Cloudflare.RecordType + `", "name": "` + config.Cloudflare.Record + `", "content": "` + currentIp +  `", "ttl": ` + config.Cloudflare.Ttl + `, "proxied":false}`)
      client.UpdateRecord(config.Cloudflare.RecordId, data)
      log.Print("Updated IP address from " + cfIp + " to " + currentIp)
    }

    time.Sleep(10 * time.Second)
  }
}
