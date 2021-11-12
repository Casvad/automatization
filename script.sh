#!/bin/bash
sudo yum update -y
sudo yum install git -y
sudo yum install java-1.8.0 -y
sudo git clone https://github.com/Casvad/automatization
export PORT=8080
export API_URLS=http://localhost:8089

java -cp "automatization/java/classes:automatization/java/dependency/*" co.edu.escuelaing.lb.SparkWebServer &
