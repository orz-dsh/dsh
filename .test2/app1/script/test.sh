#!/bin/sh

app_test_download() {
  dsh_curl_download "https://raw.githubusercontent.com/orz-dsh/dsh/main/LICENSE" "${DSH_APP_DIR}/../../test space dir/output/LICENSE"
}

app_test_upload() {
  dsh_curl_upload "https://raw.githubusercontent.com/orz-dsh/dsh/main/LICENSE" "${DSH_APP_DIR}/../../test space dir/output/LICENSE" --user-agent "dsh test upload"
}
