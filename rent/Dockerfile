FROM maven:3.9.8-eclipse-temurin-21

WORKDIR /app

# copy the project files
COPY ./pom.xml ./pom.xml


# build all dependencies for offline use
RUN mvn dependency:go-offline -B

# copy the src files
COPY ./src ./src

# build for release
RUN mvn clean package

RUN cp ./target/*.jar app.jar

EXPOSE 8080

ENTRYPOINT ["java", "-jar", "/app/app.jar"]