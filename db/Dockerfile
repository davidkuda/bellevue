# Script to setup the database with the correct tables
FROM debian:12-slim

# Install postgress client so we can update DB
RUN apt update; apt install postgresql-client -y

COPY db_init_script.sh script.sh

CMD ["bash", "script.sh"]