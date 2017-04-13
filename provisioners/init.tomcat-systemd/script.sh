#!/bin/sh
cat - >> /input-unit
mkdir -p /etc/service/tomcat
echo '#!/bin/sh' > /etc/service/tomcat/run
cat /input-unit | sed -n '/Environment\=\(.*\)/p' | sed 's/Environment=\(.*\)$/\1/' | sed "s/'//g" | sed -e "s/^\([^=]*\)=\(.*\)$/\1='\2'/" >> /etc/service/tomcat/run
echo 'exec setuidgid tomcat \' >> /etc/service/tomcat/run
cat /input-unit | sed -n '/ExecStart/p' | sed 's/ExecStart=\(.*\)$/\1/' >> /etc/service/tomcat/run

tar cf - Dockerfile /etc/service/tomcat/run
