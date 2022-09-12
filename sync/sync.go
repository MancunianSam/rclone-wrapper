package main

import (
  "context"
  "fmt"
  "github.com/aws/aws-lambda-go/lambda"
  _ "github.com/rclone/rclone/backend/drive"
  _ "github.com/rclone/rclone/backend/s3"
  "github.com/rclone/rclone/fs"
  "github.com/rclone/rclone/fs/config"
  "github.com/rclone/rclone/fs/operations"
  "github.com/rclone/rclone/fs/rc"
  "log"
)

type TransferEvent struct {
  Token       string `json:"token"`
  Bucket      string `json:"bucket"`
  DriveFile   string `json:"driveFolder"`
  DriveParent string `json:"driveParent"`
  S3Path      string `json:"s3Path"`
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

  fsDrive, driveRemoteFsErr := fs.NewFs(ctx, fmt.Sprintf("google:/%s", event.DriveParent))
  checkError(driveRemoteFsErr)

  fsAWS, awsRemoteFsErr := fs.NewFs(ctx, fmt.Sprintf("aws:%s/%s", event.Bucket, ""))
  checkError(awsRemoteFsErr)

  //Replace this with a call to copy
  copyErr := operations.CopyFile(ctx, fsAWS, fsDrive, event.S3Path, event.DriveFile)

  checkError(copyErr)

  return Response{Message: "ok"}, nil
}

func checkError(err error) {
  if err != nil {
    log.Fatal(err)
  }
}

func main() {
  //HandleRequest(TransferEvent{Token: os.Args[1], S3Path: "testpath", Bucket: "sam-test-bucket-sandbox", DriveParent: "/Test", DriveFile: "Nested/TestDoc.docx"})
  lambda.Start(HandleRequest)
}
