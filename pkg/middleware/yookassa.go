package middleware

import (
	"net"

	"github.com/gofiber/fiber/v3"
)

var allowedCIDRs = []string{
	"185.71.76.0/27",
	"185.71.77.0/27",
	"77.75.153.0/25",
	"77.75.154.128/25",
	"2a02:5180::/32",
}

var allowedIPs = []string{
	"77.75.156.11",
	"77.75.156.35",
}

var (
	parsedCIDRs []*net.IPNet
	parsedIPs   []net.IP
)

func init() {
	for _, cidr := range allowedCIDRs {
		_, network, err := net.ParseCIDR(cidr)
		if err == nil {
			parsedCIDRs = append(parsedCIDRs, network)
		}
	}

	for _, ip := range allowedIPs {
		parsed := net.ParseIP(ip)
		if parsed != nil {
			parsedIPs = append(parsedIPs, parsed)
		}
	}
}

func YookassaIPWhitelist() fiber.Handler {
	return func(ctx fiber.Ctx) error {
		clientIP := extractClientIP(ctx)

		ip := net.ParseIP(clientIP)
		if ip == nil {
			return ctx.SendStatus(fiber.StatusForbidden)
		}

		for _, allowedIP := range parsedIPs {
			if allowedIP.Equal(ip) {
				return ctx.Next()
			}
		}

		for _, network := range parsedCIDRs {
			if network.Contains(ip) {
				return ctx.Next()
			}
		}

		return ctx.SendStatus(fiber.StatusForbidden)
	}
}

func extractClientIP(ctx fiber.Ctx) string {
	ips := ctx.IPs()
	if len(ips) > 0 {
		return ips[0]
	}

	return ctx.IP()
}
