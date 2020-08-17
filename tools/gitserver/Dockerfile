FROM ubuntu:18.04

RUN apt update 2>/dev/null &&\
  apt install -y git apache2 apache2-utils 2>/dev/null
RUN a2enmod env cgi alias rewrite
COPY ./etc/git.conf /etc/apache2/sites-available/git.conf
RUN a2dissite 000-default.conf &&\
  a2ensite git.conf &&\
  git config --system http.receivepack true &&\
  git config --system http.uploadpack true
ENV APACHE_RUN_USER www-data
ENV APACHE_RUN_GROUP www-data
ENV APACHE_LOG_DIR /var/log/apache2
ENV APACHE_LOCK_DIR /var/lock/apache2
ENV APACHE_PID_FILE /var/run/apache2.pid

COPY ./etc/init_repos.sh /tmp/init_repos.sh
COPY ./repos /tmp/repos
RUN /tmp/init_repos.sh

CMD /usr/sbin/apache2ctl -D FOREGROUND

EXPOSE 80/tcp