#!/usr/bin/env sh

set -o errexit
set -o nounset
set -o pipefail

config_path="config/config"

terraform init

if [ ! -f ${config_path} ];then
  touch ${config_path}
fi

echo "pod started.." 

if [ "$(sha256sum ${config_path} |cut -d ' ' -f 1)" = "$(sha256sum $TFVARS|cut -d ' ' -f 1)" ];then
  printf "terraform infra is in sync.\n---\n"
fi

while true;do
  if [ "$(sha256sum ${config_path} |cut -d ' ' -f 1)" != "$(sha256sum $TFVARS|cut -d ' ' -f 1)" ];then
    echo "Out of sync. Variable file is updated. Applying new changes."
    terraform apply --var-file=$TFVARS -auto-approve

    if [ $? -eq 0 ]; then
      cp $TFVARS ${config_path}
      echo New configuration applied successfully.
      printf "terraform infra is in sync.\n---\n"
    else
      retry "terraform apply --var-file=$TFVARS -auto-approve"
    fi

  fi

  # config sync interval period added
  sleep 5
done

retry(){
  max_retries=3
  retry_intial_delay=2

  retry_count=0
  delay=$retry_intial_delay

  while [ $retry_count -lt $max_retries ]; do
      command=$1
      echo Executing: $command

      $command

      if [ $? -eq 0 ]; then
          echo "Command executed successfully."
          echo ----
          break
      else
          echo "ERROR: Retrying in $delay seconds..."
          sleep $delay
          delay=$((delay*2))
          retry_count=$((retry_count+1))
      fi

      if [ $retry_count -eq $max_retries ]; then
          echo "Maximum retries reached. Exiting..."
          exit 1
      fi

  done
}
