package main

import (
  "context"
  "fmt"
  "github.com/aws/aws-lambda-go/lambda"
  _ "github.com/rclone/rclone/backend/drive"
  _ "github.com/rclone/rclone/backend/s3"
  "github.com/rclone/rclone/fs"
  "github.com/rclone/rclone/fs/config"
  "github.com/rclone/rclone/fs/rc"
  "github.com/rclone/rclone/fs/sync"
  "log"
)

type TransferEvent struct {
  Token         string `json:"token"`
  Bucket        string `json:"bucket"`
  UserId        string `json:"user_id"`
  ConsignmentId string `json:"consignment_id"`
}

type Response struct {
  Message string `json:"status"`
}

func HandleRequest(event TransferEvent) (Response, error) {
  ctx := context.Background()
  remoteOpt := config.UpdateRemoteOpt{NonInteractive: true}
  googleParams := rc.Params{"token": fmt.Sprintf("{\"access_token\":\"%s\"}", event.Token)}
  s3Params := rc.Params{"provider": "AWS", "env_auth": true, "region": "eu-west-2", "acl": "private"}

  _, driveRemoteErr := config.CreateRemote(ctx, "google", "drive", googleParams, remoteOpt)
  checkError(driveRemoteErr)

  _, awsRemoteErr := config.CreateRemote(ctx, "aws", "s3", s3Params, remoteOpt)
  checkError(awsRemoteErr)

  f, driveRemoteFsErr := fs.NewFs(ctx, "google:/")
  checkError(driveRemoteFsErr)

  fa, awsRemoteFsErr := fs.NewFs(ctx, fmt.Sprintf("aws:%s/%s/%s", event.Bucket, event.UserId, event.ConsignmentId))
  checkError(awsRemoteFsErr)

  syncErr := sync.Sync(ctx, fa, f, true)
  checkError(syncErr)

  return Response{Message: "ok"}, nil
}

func checkError(err error) {
  if err != nil {
    log.Fatal(err)
  }
}

func main() {
  //HandleRequest(TransferEvent{Token: os.Args[1], ConsignmentId: "9d8524a5-9e2d-4a01-a28f-bd93dcaeda31", UserId: "cb4fa22f-1c08-4b79-b0ab-ab21dea7443e", Bucket: "sam-test-bucket-sandbox"})
  lambda.Start(HandleRequest)
}
