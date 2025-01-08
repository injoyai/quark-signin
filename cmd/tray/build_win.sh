name="quark-signin"
GOOS=windows GOARCH=amd64 go build -v -ldflags="-H windowsgui -w -s" -o ./$name.exe
echo "$name 编译完成..."
echo "开始压缩..."
#upx -9 -k "./$name.exe"
if [ -f "./$name.ex~" ]; then
  rm "./$name.ex~"
fi
if [ -f "./$name.000" ]; then
  rm "./$name.000"
fi

echo "上传至minio"
cmd.exe /c "in upload minio ./$name.exe"

sleep 2