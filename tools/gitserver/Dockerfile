FROM alpine:3.12

RUN apk update &&\
  apk add git-daemon apache2 apache2-utils
COPY ./etc/git.conf /etc/apache2/conf.d/git.conf
COPY ./etc/httpd.conf /etc/apache2/httpd.conf
RUN git config --system http.receivepack true &&\
  git config --system http.uploadpack true &&\
  git config --global user.email "gitserver@kyma-project.io" &&\
  git config --global user.name "Git Server"
ENV APACHE_LOG_DIR /var/log/apache2
ENV APACHE_LOCK_DIR /var/lock/apache2
ENV APACHE_PID_FILE /var/run/apache2.pid

COPY ./etc/init_repos.sh /tmp/init_repos.sh
COPY ./repos /tmp/repos
RUN /tmp/init_repos.sh

CMD /usr/sbin/httpd -D FOREGROUND -f /etc/apache2/httpd.conf

EXPOSE 80/tcp