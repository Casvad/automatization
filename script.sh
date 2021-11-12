sudo yum update -y
sudo yum install git -y
git clone https://github.com/Casvad/automatization
sudo yum install java-1.8.0
y
cd automatization/java
export PORT=8088
java -cp "classes:/dependency/*" co.edu.escuelaing.lb.SparkWebServer
