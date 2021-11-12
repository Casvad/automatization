package main

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"log"
	"os"
	"os/exec"
	"time"
)

func Execute() {

	awsSession := session.Must(session.NewSession())
	ec2Instance := ec2.New(awsSession)

	keyPair := createKeyPair(ec2Instance)
	writeKeyPair(keyPair)

	vpc := describeVpc(ec2Instance)
	sg := createSecurityGroup(ec2Instance, vpc.Vpcs[0].VpcId)

	tcpPort := int64(22)
	servicePort := int64(8080)

	authorizeSecurityGroup(ec2Instance, sg.GroupId, &tcpPort)
	authorizeSecurityGroup(ec2Instance, sg.GroupId, &servicePort)

	sn := describeSubnets(ec2Instance)

	reservation := runInstances(ec2Instance, keyPair.KeyName, sg.GroupId, sn.Subnets[0].SubnetId)
	for _, instance := range reservation.Instances {
		time.Sleep(20 * time.Second)
		connectInstance(keyPair, instance, ec2Instance)
	}
}

func connectInstance(keyPair *ec2.CreateKeyPairOutput, instance *ec2.Instance, ec2Instance *ec2.EC2) {
	describe, err := ec2Instance.DescribeInstances(&ec2.DescribeInstancesInput{InstanceIds: []*string{instance.InstanceId}})
	if err != nil {
		panic(err)
	}
	println("start publish on " + "ec2-user@" + *describe.Reservations[0].Instances[0].PublicDnsName)
	cmd := exec.Command("ssh", "-i", *keyPair.KeyName+".pem", "ec2-user@"+*describe.Reservations[0].Instances[0].PublicDnsName, "echo hola")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout
	cmd.Stdin = os.Stdin

	if err = cmd.Run(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			// terminated by Control-C so ignoring
			if exiterr.ExitCode() == 130 {
				return
			}
		}

		return
	}
}

func runInstances(ec2Instance *ec2.EC2, keyName, securityGroupId, subnetId *string) *ec2.Reservation {
	ami := "ami-01cc34ab2709337aa"
	instanceType := "t2.micro"
	instances := int64(3)

	r, err := ec2Instance.RunInstances(&ec2.RunInstancesInput{
		ImageId:          &ami,
		InstanceType:     &instanceType,
		KeyName:          keyName,
		MaxCount:         &instances,
		MinCount:         &instances,
		SecurityGroupIds: []*string{securityGroupId},
		SubnetId:         subnetId,
		UserData:         &ami,
	})

	if err != nil {
		panic(err)
	}

	return r
}

func describeVpc(ec2Instance *ec2.EC2) *ec2.DescribeVpcsOutput {

	vpc, err := ec2Instance.DescribeVpcs(&ec2.DescribeVpcsInput{})
	if err != nil {
		panic(err)
	}

	return vpc
}

func describeSubnets(ec2Instance *ec2.EC2) *ec2.DescribeSubnetsOutput {

	subnet, err := ec2Instance.DescribeSubnets(&ec2.DescribeSubnetsInput{})
	if err != nil {
		panic(err)
	}

	return subnet
}

func writeKeyPair(keyPair *ec2.CreateKeyPairOutput) {

	f, err := os.Create(*keyPair.KeyName + ".pem")

	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	_, err2 := f.WriteString(*keyPair.KeyMaterial)

	if err2 != nil {
		log.Fatal(err2)
	}
}

func createKeyPair(ec2Instance *ec2.EC2) *ec2.CreateKeyPairOutput {
	name := "auto-key"
	keyPair, err := ec2Instance.CreateKeyPair(&ec2.CreateKeyPairInput{
		KeyName: &name,
	})

	if err != nil {
		panic(err)
	}

	return keyPair
}

func createSecurityGroup(ec2Instance *ec2.EC2, vpcId *string) *ec2.CreateSecurityGroupOutput {
	description := "automated security group"
	groupName := "automated-sg"

	sg, err := ec2Instance.CreateSecurityGroup(&ec2.CreateSecurityGroupInput{
		Description: &description,
		GroupName:   &groupName,
		VpcId:       vpcId,
	})

	if err != nil {
		panic(err)
	}

	return sg
}

func authorizeSecurityGroup(ec2Instance *ec2.EC2, groupId *string, port *int64) {
	protocol := "tcp"
	ciDr := "0.0.0.0/0"

	_, err := ec2Instance.AuthorizeSecurityGroupIngress(&ec2.AuthorizeSecurityGroupIngressInput{
		CidrIp:     &ciDr,
		FromPort:   port,
		GroupId:    groupId,
		IpProtocol: &protocol,
		ToPort:     port,
	})

	if err != nil {
		panic(err)
	}
}
