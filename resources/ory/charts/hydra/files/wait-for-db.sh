# DSN is always in the form DSN=DB_TYPE://DB_USER:PASSWORD@DB_URL/DB_NAME?sslmode=disable
# But DSN for mysql is in the form DSN=DB_TYPE://DB_USER:PASSWORD@tcp(DB_URL)/DB_NAME?parseTime=true

# Extract DB_TYPE to check if the DB_TYPE is mysql
DB_TYPE_VALUE=$(echo $DSN | awk -F '://' '{print $1}')

if [ $DB_TYPE_VALUE == "mysql" ]; then
  # Extract DB_URL by cutting between @ and first / and extracting the string between brackets
  DB_URL_PORT=$(echo $DSN | cut -d '@' -f2 | cut -d '/' -f 1 | cut -d "(" -f2 | cut -d ")" -f1)
else
  # Extract DB_URL by cutting between @ and first /
  DB_URL_PORT=$(echo $DSN | cut -d '@' -f2 | cut -d '/' -f 1 )
fi

# DB_URL is expected to be mydb.mynamespace.svc.cluster.local:1234, but the port can be optional
# Check if it given by looking for :
if echo "$DB_URL_PORT" | grep -q ':'; then
  # Split URL and port by :
  DB_URL=$(echo $DB_URL_PORT | cut -d ':' -f 1)
  DB_PORT=$(echo $DB_URL_PORT | cut -d ':' -f 2)
else
  # Use the full url, since no port was given
  DB_URL=$DB_URL_PORT
fi

# If port was given, we use it, if not, empty var is expanded to 0
until nc -zv -w 5 $DB_URL $DB_PORT; do
  echo "$DB_URL not yet ready"
  sleep 5
done