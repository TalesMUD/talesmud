echo "--- Backend:"
find . -name '*.go' | xargs wc -l
echo "--- Frontend:"
find . -path ./public/app/node_modules -prune -name '*.js' -or -name '*.svelte' | xargs wc -l
