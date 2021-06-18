package mysql

/*

import (
	. "godata"
)

func TestCoffeeDatabase(t *testing.T) {

	cxnParams := &MySQLConnectionParams{
		Hostname: "localhost",
		Port:     "3306",
		Database: "Coffee",
		Username: "dev",
		Password: "dev",
	}
	provider := BuildMySQLProvider(cxnParams, "Coffee")

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

	actual, err := provider.BuildMetadata().String()

	expected := "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n"
	expected += "<edmx:Edmx xmlns:edmx=\"http://docs.oasis-open.org/odata/ns/edmx\" Version=\"4.0\">\n"
	expected += "    <edmx:DataServices>\n"
	expected += "        <Schema Namespace=\"Coffee\">\n"
	expected += "            <EntityContainer Name=\"Coffee\">\n"
	expected += "                <EntitySet Name=\"Varieties\" EntityType=\"Coffee.Variety\">\n"
	expected += "                    <NavigationPropertyBinding Path=\"Beans\" Target=\"Beans\"></NavigationPropertyBinding>\n"
	expected += "                </EntitySet>\n"
	expected += "                <EntitySet Name=\"RoastLevels\" EntityType=\"Coffee.RoastLevel\">\n"
	expected += "                    <NavigationPropertyBinding Path=\"Beans\" Target=\"Beans\"></NavigationPropertyBinding>\n"
	expected += "                </EntitySet>\n"
	expected += "                <EntitySet Name=\"Processes\" EntityType=\"Coffee.Process\">\n"
	expected += "                    <NavigationPropertyBinding Path=\"Beans\" Target=\"Beans\"></NavigationPropertyBinding>\n"
	expected += "                </EntitySet>\n"
	expected += "                <EntitySet Name=\"Beans\" EntityType=\"Coffee.Bean\">\n"
	expected += "                    <NavigationPropertyBinding Path=\"Roaster\" Target=\"Roaster\"></NavigationPropertyBinding>\n"
	expected += "                    <NavigationPropertyBinding Path=\"RoastLevel\" Target=\"RoastLevel\"></NavigationPropertyBinding>\n"
	expected += "                    <NavigationPropertyBinding Path=\"Process\" Target=\"Process\"></NavigationPropertyBinding>\n"
	expected += "                    <NavigationPropertyBinding Path=\"Varieties\" Target=\"Varieties\"></NavigationPropertyBinding>\n"
	expected += "                </EntitySet>\n"
	expected += "                <EntitySet Name=\"Roasters\" EntityType=\"Coffee.Roaster\">\n"
	expected += "                    <NavigationPropertyBinding Path=\"Beans\" Target=\"Beans\"></NavigationPropertyBinding>\n"
	expected += "                </EntitySet>\n"
	expected += "            </EntityContainer>\n"
	expected += "            <EntityType Name=\"Roaster\">\n"
	expected += "                <Key>\n"
	expected += "                    <PropertyRef Name=\"RoasterID\"></PropertyRef>\n"
	expected += "                </Key>\n"
	expected += "                <Property Name=\"RoasterID\" Type=\"Edm.Int32\"></Property>\n"
	expected += "                <Property Name=\"Name\" Type=\"Edm.String\"></Property>\n"
	expected += "                <Property Name=\"Location\" Type=\"Edm.String\"></Property>\n"
	expected += "                <Property Name=\"Website\" Type=\"Edm.String\"></Property>\n"
	expected += "                <NavigationProperty Name=\"Beans\" Type=\"Collection(Coffee.Bean)\" Partner=\"Roaster\">\n"
	expected += "                    <ReferentialConstraint Property=\"RoasterID\" ReferencedProperty=\"RoasterID\"></ReferentialConstraint>\n"
	expected += "                </NavigationProperty>\n"
	expected += "            </EntityType>\n"
	expected += "            <EntityType Name=\"Variety\">\n"
	expected += "                <Key>\n"
	expected += "                    <PropertyRef Name=\"VarietyID\"></PropertyRef>\n"
	expected += "                </Key>\n"
	expected += "                <Property Name=\"VarietyID\" Type=\"Edm.Int32\"></Property>\n"
	expected += "                <Property Name=\"Name\" Type=\"Edm.String\"></Property>\n"
	expected += "                <NavigationProperty Name=\"Beans\" Type=\"Collection(Coffee.Bean)\" Partner=\"Varieties\"></NavigationProperty>\n"
	expected += "            </EntityType>\n"
	expected += "            <EntityType Name=\"RoastLevel\">\n"
	expected += "                <Key>\n"
	expected += "                    <PropertyRef Name=\"RoastLevelID\"></PropertyRef>\n"
	expected += "                </Key>\n"
	expected += "                <Property Name=\"RoastLevelID\" Type=\"Edm.Int32\"></Property>\n"
	expected += "                <Property Name=\"Order\" Type=\"Edm.Int32\"></Property>\n"
	expected += "                <Property Name=\"Name\" Type=\"Edm.String\"></Property>\n"
	expected += "                <Property Name=\"Qualifier\" Type=\"Edm.String\"></Property>\n"
	expected += "                <NavigationProperty Name=\"Beans\" Type=\"Collection(Coffee.Bean)\" Partner=\"RoastLevel\">\n"
	expected += "                    <ReferentialConstraint Property=\"RoastLevelID\" ReferencedProperty=\"RoastLevelID\"></ReferentialConstraint>\n"
	expected += "                </NavigationProperty>\n"
	expected += "            </EntityType>\n"
	expected += "            <EntityType Name=\"Process\">\n"
	expected += "                <Key>\n"
	expected += "                    <PropertyRef Name=\"ProcessID\"></PropertyRef>\n"
	expected += "                </Key>\n"
	expected += "                <Property Name=\"ProcessID\" Type=\"Edm.Int32\"></Property>\n"
	expected += "                <Property Name=\"Name\" Type=\"Edm.String\"></Property>\n"
	expected += "                <NavigationProperty Name=\"Beans\" Type=\"Collection(Coffee.Bean)\" Partner=\"Process\">\n"
	expected += "                    <ReferentialConstraint Property=\"ProcessID\" ReferencedProperty=\"ProcessID\"></ReferentialConstraint>\n"
	expected += "                </NavigationProperty>\n"
	expected += "            </EntityType>\n"
	expected += "            <EntityType Name=\"Bean\">\n"
	expected += "                <Key>\n"
	expected += "                    <PropertyRef Name=\"BeanID\"></PropertyRef>\n"
	expected += "                </Key>\n"
	expected += "                <Property Name=\"BeanID\" Type=\"Edm.Int32\"></Property>\n"
	expected += "                <Property Name=\"Country\" Type=\"Edm.String\"></Property>\n"
	expected += "                <Property Name=\"Region\" Type=\"Edm.String\"></Property>\n"
	expected += "                <Property Name=\"MinElevation\" Type=\"Edm.Int32\"></Property>\n"
	expected += "                <Property Name=\"MaxElevation\" Type=\"Edm.Int32\"></Property>\n"
	expected += "                <Property Name=\"RoasterID\" Type=\"Edm.Int32\"></Property>\n"
	expected += "                <Property Name=\"RoastLevelID\" Type=\"Edm.Int32\"></Property>\n"
	expected += "                <Property Name=\"ProcessID\" Type=\"Edm.Int32\"></Property>\n"
	expected += "                <NavigationProperty Name=\"Roaster\" Type=\"Coffee.Roaster\" Partner=\"Beans\"></NavigationProperty>\n"
	expected += "                <NavigationProperty Name=\"RoastLevel\" Type=\"Coffee.RoastLevel\" Partner=\"Beans\"></NavigationProperty>\n"
	expected += "                <NavigationProperty Name=\"Process\" Type=\"Coffee.Process\" Partner=\"Beans\"></NavigationProperty>\n"
	expected += "                <NavigationProperty Name=\"Varieties\" Type=\"Collection(Coffee.Variety)\" Partner=\"Beans\"></NavigationProperty>\n"
	expected += "            </EntityType>\n"
	expected += "        </Schema>\n"
	expected += "    </edmx:DataServices>\n"
	expected += "</edmx:Edmx>"

	if err != nil {
		t.Error(err)
	}

	if actual != expected {
		t.Error("Expected: \n"+expected, "\n\nGot: \n"+actual)
	}

}

func BenchmarkCoffeeDatabase(b *testing.B) {
	for i := 0; i < b.N; i++ {

		cxnParams := &MySQLConnectionParams{
			Hostname: "localhost",
			Port:     "3306",
			Database: "Coffee",
			Username: "dev",
			Password: "dev",
		}
		provider := BuildMySQLProvider(cxnParams, "Coffee")

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

		provider.BuildMetadata()
	}
}
*/
