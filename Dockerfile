From hashicorp/terraform:1.7.0

WORKDIR /kubernetes-pki
COPY providers.tf main.tf variables.tf backend.tf script.sh generate_varfile_from_crd.sh /kubernetes-pki/

RUN apk add jq
RUN apk add curl

# Removes local kube config; use service account  
RUN sed -i '/config_path/ d' /kubernetes-pki/providers.tf
RUN sed -i '/kube_config_path/{N;N;d}' /kubernetes-pki/variables.tf
RUN chmod +x script.sh

#TODO: Update later
ENTRYPOINT ["./script.sh"]