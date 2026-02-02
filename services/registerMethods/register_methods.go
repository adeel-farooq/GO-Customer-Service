package registerMethods

import (
	"go-cloud-customer/controllers/authmodule"
	"go-cloud-customer/controllers/customer"
	"go-cloud-customer/controllers/restricted"
	"go-cloud-customer/services"
	"log"
	"net/rpc"
)

func RegisterCustomerServices(rpcServer *rpc.Server) {
	/* Register health check service */
	err := rpcServer.Register(new(services.HealthCheckService))
	if err != nil {
		log.Fatalf("Error registering health check service: %v", err)
	}

	/* Register customer services */
	err = rpcServer.Register(new(customer.CustomerService))
	if err != nil {
		log.Fatalf("Error registering customer service: %v", err)
	}

	/* Register restricted services */
	err = rpcServer.Register(new(restricted.RestrictedService))
	if err != nil {
		log.Fatalf("Error registering restricted service: %v", err)
	}

	/* Register auth module services */
	err = rpcServer.Register(new(authmodule.AuthModuleService))
	if err != nil {
		log.Fatalf("Error registering authmodule service: %v", err)
	}
	// err = rpcServer.Register(new(restricted.RestrictedService))
	// if err != nil {
	// 	log.Fatalf("Error registering restricted service: %v", err)
	// }

	// // Compatibility alias: some callers use request type name as service name.
	// err = rpcServer.RegisterName("GetErrorMessagesRequest", new(restricted.RestrictedService))
	// if err != nil {
	// 	log.Fatalf("Error registering restricted alias service: %v", err)
	// }

}
