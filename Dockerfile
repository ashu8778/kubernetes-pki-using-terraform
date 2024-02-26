#Initial Dockerfile
From hashicorp/terraform:1.7.0

RUN mkdir /kubernetes-pki
WORKDIR /kubernetes-pki
COPY . /kubernetes-pki/

#TODO: Update later
ENTRYPOINT ["sleep", "infinite"]