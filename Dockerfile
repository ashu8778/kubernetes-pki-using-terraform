From hashicorp/terraform:1.7.0

RUN mkdir /kubernetes-pki
WORKDIR /kubernetes-pki
COPY providers.tf main.tf variables.tf  /kubernetes-pki/

# Removes local kube config; use service account  
RUN sed -i '/config_path/ d' /kubernetes-pki/providers.tf

#TODO: Update later
ENTRYPOINT ["sleep", "infinite"]