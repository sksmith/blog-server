# Start the blog server
blog-server
status=$?
if [ $status -ne 0 ]; then
  echo "Failed to start blog-server: $status"
  exit $status
fi

while sleep 60; do
  ps aux |grep blog-server |grep -q -v grep
  BLOG_SERVER_STATUS=$?
  if [ $BLOG_SERVER_STATUS -ne 0 ]; then
    echo "Blog Server exited."
    exit 1
  fi
done