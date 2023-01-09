package helpers

import (
	"fmt"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"log"
	"net"
)

// GetLoadBalancerIp returns the IP of the load balancer from the load balancer ingress object
func GetLoadBalancerIp(loadBalancerIngress map[string]interface{}) (net.IP, error) {
	loadBalancerIP, err := getIpBasedLoadBalancerIp(loadBalancerIngress)

	if err == nil {
		return loadBalancerIP, nil
	} else {
		log.Printf("Falling back to reading DNS based load balancer IP, because of: %s\n", err)
		return getDnsBasedLoadBalancerIp(loadBalancerIngress)
	}
}

func getIpBasedLoadBalancerIp(lbIngress map[string]interface{}) (net.IP, error) {
	loadBalancerIP, found, err := unstructured.NestedString(lbIngress, "ip")

	if err != nil || !found {
		return nil, fmt.Errorf("could not get IP based load balancer IP: %s", err)
	}

	ip := net.ParseIP(loadBalancerIP)
	if ip == nil {
		return nil, fmt.Errorf("failed to parse IP from load balancer IP %s", loadBalancerIP)
	}

	return ip, nil
}

func getDnsBasedLoadBalancerIp(lbIngress map[string]interface{}) (net.IP, error) {
	loadBalancerHostname, found, err := unstructured.NestedString(lbIngress, "hostname")

	if err != nil || !found {
		return nil, fmt.Errorf("could not get DNS based load balancer IP: %s", err)
	}

	ips, err := net.LookupIP(loadBalancerHostname)
	if err != nil || len(ips) < 1 {
		return nil, fmt.Errorf("could not get IPs by load balancer hostname: %s", err)
	}

	return ips[0], nil
}
