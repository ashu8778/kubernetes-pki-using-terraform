#!/bin/sh

hash="$(sha256sum $TFVARS|cut -d ' ' -f 1)"
terraform apply --var-file=$TFVARS -auto-approve
if [ $? -ne 0 ]; then
  retry "terraform apply --var-file=$TFVARS -auto-approve"
fi

while true;do
  if [ "$hash" != "$(sha256sum $TFVARS|cut -d ' ' -f 1)" ];then
    echo "Variable file is updated. Applying new changes."
    terraform apply --var-file=$TFVARS -auto-approve
    if [ $? -ne 0 ]; then
      retry "terraform apply --var-file=$TFVARS -auto-approve"
    fi
    hash="$(sha256sum $TFVARS|cut -d ' ' -f 1)"
  fi
  # interval period added
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
