package servers

var (
	backendServiceNames = []string{
		basketService,
		contentService,
		customerService,
		identityService,
		navigationService,
		orderService,
		priceService,
		productService,
		searchService,
		shipppingService,
	}

	homePageServices = map[string]bool{
		navigationService: true,
		contentService:    true,
		searchService:     true,
		productService:    true,
		priceService:      true,
		customerService:   true,
		basketService:     true,
	}
	productListingServices = map[string]bool{
		navigationService: true,
		contentService:    true,
		searchService:     true,
		productService:    true,
		priceService:      true,
		customerService:   true,
		basketService:     true,
	}
	productDetailServices = map[string]bool{
		navigationService: true,
		contentService:    true,
		searchService:     true,
		productService:    true,
		priceService:      true,
		customerService:   true,
		basketService:     true,
	}
	categoryListingServices = map[string]bool{
		navigationService: true,
		contentService:    true,
		searchService:     true,
		productService:    true,
		priceService:      true,
		customerService:   true,
		basketService:     true,
	}
	categoryDetailServices = map[string]bool{
		navigationService: true,
		contentService:    true,
		searchService:     true,
		productService:    true,
		priceService:      true,
		customerService:   true,
		basketService:     true,
	}
	searchServices = map[string]bool{
		navigationService: true,
		contentService:    true,
		searchService:     true,
		productService:    true,
		priceService:      true,
		customerService:   true,
		identityService:   true,
	}
	accountServices = map[string]bool{
		navigationService: true,
		contentService:    true,
		searchService:     true,
		productService:    true,
		priceService:      true,
		customerService:   true,
		identityService:   true,
	}
	checkoutServices = map[string]bool{
		navigationService: true,
		contentService:    true,
		searchService:     true,
		productService:    true,
		priceService:      true,
		customerService:   true,
		basketService:     true,
	}
)

// Backend represents a backend service
type Backend struct {
	Address string
	Name    string
}
