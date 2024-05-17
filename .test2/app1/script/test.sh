#!/bin/sh

app_test_download() {
  dsh_curl_download "https://raw.githubusercontent.com/orz-dsh/dsh/main/LICENSE" "${PWD}/.test2/test space dir/output/LICENSE"
}

app_test_upload() {
  dsh_curl_upload "https://raw.githubusercontent.com/orz-dsh/dsh/main/LICENSE" "${PWD}/.test2/test space dir/output/LICENSE" --user-agent "dsh test upload"
}
