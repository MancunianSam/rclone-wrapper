package main

import (
  "context"
  "encoding/json"
  "fmt"
  "github.com/aws/aws-lambda-go/events"
  "github.com/aws/aws-lambda-go/lambda"
  _ "github.com/rclone/rclone/backend/drive"
  "github.com/rclone/rclone/cmd"
  "github.com/rclone/rclone/fs/config"
  "github.com/rclone/rclone/fs/operations"
  "github.com/rclone/rclone/fs/rc"
  "log"
  "net/http"
)

type MetadataItems struct {
  Items []Metadata `json:"items"`
}
type Metadata struct {
  Path  string `json:"token"`
  Name  string `json:"name"`
  Size  int64  `json:"size"`
  IsDir bool   `json:"is_dir"`
  Hash  string `json:"hash"`
}

type MetadataEvent struct {
  Token string `json:"token"`
}

func checkError(err error) {
  if err != nil {
    log.Fatal(err)
  }
}

func HandleRequest(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
  ctx := context.Background()
  metadataEvent := MetadataEvent{}
  unmarshalErr := json.Unmarshal([]byte(req.Body), &metadataEvent)
  checkError(unmarshalErr)
  remoteOpt := config.UpdateRemoteOpt{NonInteractive: true}
  googleParams := rc.Params{"token": fmt.Sprintf("{\"access_token\":\"%s\"}", metadataEvent.Token)}

  _, driveRemoteErr := config.CreateRemote(ctx, "google", "drive", googleParams, remoteOpt)
  checkError(driveRemoteErr)

  fsrc := cmd.NewFsSrc([]string{"google:/Test"})

  opts := operations.ListJSONOpt{ShowHash: true, Recurse: true}
  items := MetadataItems{}

  err := operations.ListJSON(ctx, fsrc, "", &opts, func(item *operations.ListJSONItem) error {
    metadata := Metadata{
      Path:  item.Path,
      Name:  item.Name,
      Size:  item.Size,
      IsDir: item.IsDir,
      Hash:  item.Hashes["md5"],
    }
    newItems := append(items.Items, metadata)
    items.Items = newItems
    return nil
  })
  checkError(err)
  bytes, marshallError := json.Marshal(items)

  checkError(marshallError)

  return events.APIGatewayProxyResponse{
    StatusCode: http.StatusOK,
    Body:       string(bytes),
    Headers:    map[string]string{"Content-Type": "application/json"},
  }, nil
}

func main() {
  lambda.Start(HandleRequest)
}
