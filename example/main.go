package example

/*
import (
	. "godata"
)
*/

func HelloWorld() {

}

func CacheMiddleware() {

}

func AuthorizationMiddleware() {

}

func main() {
	/*
		provider := &MySQLGoDataProvider{
			Hostname: "localhost",
			Port:     "3306",
			Database: "Coffee",
			Username: "dev",
			Password: "dev",
		}

		roaster := provider.ExposeEntity("roaster", "Roaster")
		roaster.ExposeKey("id", "RoasterID", GoDataInt32)
		roaster.ExposePrimitive("name", "Name", GoDataString)
		roaster.ExposePrimitive("location", "Location", GoDataString)
		roaster.ExposePrimitive("website", "Website", GoDataString)
		roasterSet := provider.ExposeEntitySet(roaster, "Roasters")

		variety := provider.ExposeEntity("variety", "Variety")
		variety.ExposeKey("id", "VarietyID", GoDataInt32)
		variety.ExposePrimitive("name", "Name", GoDataString)
		varietySet := provider.ExposeEntitySet(variety, "Varieties")

		roastLevel := provider.ExposeEntity("roast_level", "RoastLevel")
		roastLevel.ExposeKey("id", "RoastLevelID", GoDataInt32)
		roastLevel.ExposePrimitive("order", "Order", GoDataInt32)
		roastLevel.ExposePrimitive("name", "Name", GoDataString)
		roastLevel.ExposePrimitive("qualifier", "Qualifier", GoDataString)
		roastLevelSet := provider.ExposeEntitySet(roastLevel, "RoastLevels")

		process := provider.ExposeEntity("process", "Process")
		process.ExposeKey("id", "ProcessID", GoDataInt32)
		process.ExposePrimitive("name", "Name", GoDataString)
		processSet := provider.ExposeEntitySet(process, "Processes")

		bean := provider.ExposeEntity("bean", "Bean")
		bean.ExposeKey("id", "BeanID", GoDataInt32)
		bean.ExposePrimitive("country", "Country", GoDataString)
		bean.ExposePrimitive("region", "Region", GoDataString)
		bean.ExposePrimitive("min_elevation", "MinElevation", GoDataInt32)
		bean.ExposePrimitive("max_elevation", "MaxElevation", GoDataInt32)
		beanSet := provider.ExposeEntitySet(bean, "Beans")

		// map many beans to one roaster
		provider.ExposeManyToOne(bean, roaster, "roaster_id", "Roaster", "Beans")
		provider.ExposeManyToOne(bean, roastLevel, "roaster_level_id", "RoastLevel", "Beans")
		provider.ExposeManyToOne(bean, process, "process_id", "Process", "Beans")
		provider.ExposeManyToMany(bean, variety, "bean_variety_map", "Varieties", "Beans")
		provider.BindProperty(beanSet, roasterSet, "Roaster", "Roaster", "Beans", "Beans")
		provider.BindProperty(beanSet, roastLevelSet, "RoastLevel", "RoastLevel", "Beans", "Beans")
		provider.BindProperty(beanSet, processSet, "Process", "Process", "Beans", "Beans")
		provider.BindProperty(beanSet, varietySet, "Varieties", "Varieties", "Beans", "Beans")

		service := BuildService(provider)
		service.ListenAndServe(":8080", "http://localhost")
	*/

	//service.AttachMiddleware(CacheMiddleware)
	//service.AttachMiddleware(AuthorizationMiddleware)
	//service.BindAction(HelloWorld)
	//service.BindFunction(HelloWorld)
}
