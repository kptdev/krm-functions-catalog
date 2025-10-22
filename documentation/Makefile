serve: clean copy-examples inject-relatedexamples
	hugo server \
		--disableFastRender \
		--buildDrafts \
		--buildFuture \
		--ignoreCache
		--printI18nWarnings \
		--printMemoryUsage \
		--printPathWarnings \
		--printUnusedTemplates \
		--templateMetrics \
		--templateMetricsHints \
		--gc

production-build: copy-examples inject-relatedexamples
	hugo version
	hugo \
		--minify
	npx -y pagefind --site public

preview-build: copy-examples inject-relatedexamples
	hugo \
		--baseURL $(DEPLOY_PRIME_URL) \
		--buildDrafts \
		--buildFuture \
		--minify
	npx -y pagefind --site public

copy-examples:
	@mkdir -p content/en/examples
	@echo "Syncing only README.md files"
	@rsync -av \
		--exclude='.expected/' \
		--include='*/' \
		--include='README.md' \
		--exclude='*' \
		../examples/ content/en/examples/

inject-relatedexamples:
	@echo "Injecting {{< relatedexamples >}} shortcode above 'Overview'..."
	@find content/en/examples -type f -name "README.md" ! -path "*/.expected/*" | while read file; do \
		if ! grep -q '{{< relatedexamples' "$$file"; then \
			if grep -qi '^### *Overview' "$$file"; then \
				echo "Updating $$file with relatedexamples shortcode"; \
				tmpfile=$$(mktemp); \
				awk -v shortcode="{{< relatedexamples >}}" 'BEGIN{inserted=0} \
					/^### *[Oo]verview/ && !inserted {print shortcode "\n"; inserted=1} \
					{print}' "$$file" > "$$tmpfile"; \
				mv "$$tmpfile" "$$file"; \
			fi; \
		fi; \
	done


clean:
	rm -rf content/en/examples \
	       node_modules \
	       public \
	       resources \
	       .hugo_build.lock \
	       package-lock.json