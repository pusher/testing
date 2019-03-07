update-config: 
	#TODO: gcloud get-credentials
	kubectl create configmap config --from-file=config=prow/config.yaml --dry-run -o yaml | kubectl --namespace=default replace configmap config -f -
	kubectl create configmap plugins --from-file=plugins=prow/plugins.yaml --dry-run -o yaml | kubectl --namespace=default replace configmap plugins -f -