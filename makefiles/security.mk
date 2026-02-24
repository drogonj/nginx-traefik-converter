scan/code: ## Scans code for vulnerabilities with Trivy
	@docker-compose --project-name trivy -f docker-compose.trivy.yml run --rm trivy fs /nginx-traefik-converter
