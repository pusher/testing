if command -v gsed &>/dev/null; then
  SED="gsed -E"
else
  SED="sed -E"
fi

if ! (${SED} --version 2>&1 | grep -q GNU); then
  # darwin is great (not)
  echo "!!! GNU sed is required.  If on OS X, use 'brew install gnu-sed'." >&2
  exit 1
fi

echo "Preflight checks completed, ready for liftoff!"
