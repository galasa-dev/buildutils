FROM harbor-cicsk8s.hursley.ibm.com/dockerhub/library/ubuntu:latest

ENV DEBIAN_FRONTEND="noninteractive" TZ="Europe/London"
RUN apt-get update && apt-get install -y unzip \
    curl \
    xz-utils \
    golang \
    make \
    python2 \
    gpg && apt-get clean && rm -rf /var/lib/apt/lists/*

# Installing Java 11
#RUN mkdir /usr/local/java
#RUN curl -L https://github.com/AdoptOpenJDK/openjdk11-binaries/releases/download/jdk11u-2021-05-07-07-34/OpenJDK11U-jdk_x64_linux_openj9_2021-05-07-07-34.tar.gz \
#       -o /usr/local/java/OpenJDK11U-jdk_x64_linux_openj9_2021-05-07-07-34.tar.gz \ 
#      && tar -zxvf /usr/local/java/OpenJDK11U-jdk_x64_linux_openj9_2021-05-07-07-34.tar.gz -C /usr/local/java \
#      && rm /usr/local/java/OpenJDK11U-jdk_x64_linux_openj9_2021-05-07-07-34.tar.gz 
#ENV JAVA_HOME=/usr/local/java/jdk-11.0.12+1
#ENV PATH=$PATH:$JAVA_HOME/bin

# Installing Java 8
RUN mkdir /usr/local/java
RUN curl -L https://github.com/AdoptOpenJDK/openjdk8-binaries/releases/download/jdk8u292-b10/OpenJDK8U-jdk_x64_linux_hotspot_8u292b10.tar.gz \
       -o /usr/local/java/openjdk.tar.gz \
      && tar -zxvf /usr/local/java/openjdk.tar.gz -C /usr/local/java \
      && rm /usr/local/java/openjdk.tar.gz
ENV JAVA_HOME=/usr/local/java/jdk8u292-b10
ENV PATH=$PATH:$JAVA_HOME/bin

COPY ibmca.pem /usr/local/java
COPY ibminter.pem /usr/local/java
RUN keytool -import -trustcacerts -alias ibmca -file /usr/local/java/ibmca.pem -keystore $JAVA_HOME/jre/lib/security/cacerts -storepass changeit -noprompt
RUN keytool -import -trustcacerts -alias ibminter -file /usr/local/java/ibminter.pem -keystore $JAVA_HOME/jre/lib/security/cacerts  -storepass changeit -noprompt

RUN  cat /usr/local/java/ibmca.pem >> /etc/ssl/certs/ca-certificates.crt
RUN  cat /usr/local/java/ibminter.pem >> /etc/ssl/certs/ca-certificates.crt

RUN java -version

# Installing maven
RUN mkdir /opt/maven
RUN curl -L https://downloads.apache.org/maven/maven-3/3.8.6/binaries/apache-maven-3.8.6-bin.tar.gz -o /opt/maven/apache-maven-3.8.6-bin.tar.gz && tar -zxvf /opt/maven/apache-maven-3.8.6-bin.tar.gz -C /opt/maven/ && rm /opt/maven/apache-maven-3.8.6-bin.tar.gz
ENV PATH=$PATH:/opt/maven/apache-maven-3.8.6/bin

RUN mvn -v

# Installing Github CLI
RUN curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg | gpg --dearmor -o /usr/share/keyrings/githubcli-archive-keyring.gpg
RUN echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" | tee /etc/apt/sources.list.d/github-cli.list > /dev/null
RUN apt update && apt install -y git \
    gh && apt-get clean && rm -rf /var/lib/apt/lists/*

# Installing 6.8.2 version of gradle
RUN mkdir /opt/gradle
RUN curl -L https://services.gradle.org/distributions/gradle-6.8.2-bin.zip -o /tmp/gradle-6.8.2-bin.zip && unzip -d /opt/gradle /tmp/gradle-6.8.2-bin.zip && rm /tmp/gradle-6.8.2-bin.zip
ENV PATH=$PATH:/opt/gradle/gradle-6.8.2/bin

# Retrieving the openapi jar
RUN mkdir /opt/openapi
RUN curl -L https://repo1.maven.org/maven2/org/openapitools/openapi-generator-cli/5.1.0/openapi-generator-cli-5.1.0.jar -o /opt/openapi/openapi-generator-cli.jar

# Installing node and npm
RUN mkdir /opt/node
RUN curl -L https://nodejs.org/dist/v12.21.0/node-v12.21.0-linux-x64.tar.xz -o /opt/node/node-v12.21.0-linux-x64.tar.xz && tar xf /opt/node/node-v12.21.0-linux-x64.tar.xz -C /opt/node/ && rm /opt/node/node-v12.21.0-linux-x64.tar.xz
ENV PATH=$PATH:/opt/node/node-v12.21.0-linux-x64/bin

# Install angular cli
RUN npm install -g @angular/cli@9.1.12

# Installing kubectl
RUN mkdir -p /opt/k8s/bin
RUN curl -L https://dl.k8s.io/release/v1.20.0/bin/linux/amd64/kubectl -o /opt/k8s/bin/kubectl
RUN chmod +x /opt/k8s/bin/kubectl
ENV PATH=$PATH:/opt/k8s/bin

RUN mkdir -p /opt/maven/preload
COPY pom.xml /opt/maven/preload
WORKDIR /opt/maven/preload
RUN mvn -B deploy && rm -rf /opt/maven/preload/target && rm -rf /opt/maven/preload/repo && rm -rf /root/.m2/repository/dev/galasa

# Installing galasabld
RUN mkdir -p /opt/galasa/bin
COPY galasabld-linux-amd64 /opt/galasa/bin/galasabld
RUN chmod +x /opt/galasa/bin/galasabld
ENV PATH=$PATH:/opt/galasa/bin

RUN mkdir -p /workspace
WORKDIR /workspace
