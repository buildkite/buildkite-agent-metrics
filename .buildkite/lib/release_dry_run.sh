release_dry_run() {
  if [[ "${RELEASE_DRY_RUN:-false}" == "true" ]]; then
    echo "$@"
  else
    "$@"
  fi
}
