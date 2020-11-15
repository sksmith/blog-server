# Start gitwatch
gitwatch &
status=$?
if [ $status -ne 0 ]; then
  echo "Failed to start gitwatch: $status"
  exit $status
fi

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

  ps aux |grep gitwatch |grep -q -v grep
  GITWATCH_STATUS=$?
  if [ $GITWATCH_STATUS -ne 0 ]; then
    echo "Gitwatch process exited."
    exit 1
  fi
done