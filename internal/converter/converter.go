package converter

import (
	"fmt"
	"log"

	"github.com/kong/go-database-reconciler/pkg/file"
)

var (
	// RateLimitingPluginName is the name of the rate limiting plugin
	RateLimitingPluginName string = "rate-limiting-advanced"
)

func contains(slice []interface{}, item string) bool {
	for _, s := range slice {
		if s.(string) == item {
			return true
		}
	}
	return false
}

func removeAllPluginConsumerGroups(targetContent *file.Content) {
	// Start checking all routes
	for _, service := range targetContent.Services {
		for _, route := range service.Routes {
			for _, plugin := range route.Plugins {
				if *plugin.Name == RateLimitingPluginName {
					plugin.Config["consumer_groups"] = nil
					plugin.Config["enforce_consumer_groups"] = false
				}
			}
		}
	}

	// Then the servicesless-routes
	for _, route := range targetContent.Routes {
		for _, plugin := range route.Plugins {
			if *plugin.Name == RateLimitingPluginName {
				plugin.Config["consumer_groups"] = nil
				plugin.Config["enforce_consumer_groups"] = false
			}
		}
	}

	// Then all services
	for _, service := range targetContent.Services {
		for _, plugin := range service.Plugins {
			if *plugin.Name == RateLimitingPluginName {
				plugin.Config["consumer_groups"] = nil
				plugin.Config["enforce_consumer_groups"] = false
			}
		}
	}

	// Then grab the one from the global state
	for _, plugin := range targetContent.Plugins {
		if *plugin.Name == RateLimitingPluginName {
			plugin.Config["consumer_groups"] = nil
			plugin.Config["enforce_consumer_groups"] = false
		}
	}
}

func findNearestPlugin(groupName *string, targetContent *file.Content) *file.FPlugin {
	// Start checking all routes
	for _, service := range targetContent.Services {
		for _, route := range service.Routes {
			for _, plugin := range route.Plugins {
				if *plugin.Name == RateLimitingPluginName {
					// Check if this plugin references the consumer group
					configuredGroups, ok := plugin.Config["consumer_groups"].([]interface{})
					if !ok {
						return nil
					}

					if contains(configuredGroups, *groupName) {
						return plugin
					}
				}
			}
		}
	}

	fmt.Println("No plugin found in /services/routes")

	// Then the servicesless-routes
	for _, route := range targetContent.Routes {
		for _, plugin := range route.Plugins {
			if *plugin.Name == RateLimitingPluginName {
				// Check if this plugin references the consumer group
				configuredGroups, ok := plugin.Config["consumer_groups"].([]interface{})
				if !ok {
					return nil
				}

				if contains(configuredGroups, *groupName) {
					return plugin
				}
			}
		}
	}

	fmt.Println("No plugin found in /routes")

	// Then all services
	for _, service := range targetContent.Services {
		for _, plugin := range service.Plugins {
			if *plugin.Name == RateLimitingPluginName {
				// Check if this plugin references the consumer group
				configuredGroups, ok := plugin.Config["consumer_groups"].([]interface{})
				if !ok {
					return nil
				}

				if contains(configuredGroups, *groupName) {
					return plugin
				}
			}
		}
	}

	fmt.Println("No plugin found in /services")

	// Then grab the one from the global state
	for _, plugin := range targetContent.Plugins {
		if *plugin.Name == RateLimitingPluginName {
			// Check if this plugin references the consumer group
			configuredGroups, ok := plugin.Config["consumer_groups"].([]interface{})
			if !ok {
				return nil
			}

			if contains(configuredGroups, *groupName) {
				return &plugin
			}
		}
	}

	fmt.Println("No plugin found in /plugins")

	// Otherwise there is none for this consumer group!
	return nil
}

func Convert(source, destination string) {
	fmt.Printf("Converting %s to %s\n", source, destination)

	// 0. Read input file
	files := []string{source}
	targetContent, err := file.GetContentFromFiles(files, false)

	if err != nil {
		log.Fatalf("Error parsing file %s into Kong state object: %s\n", source, err)
	}

	// 1. For each consumer group "vague" rate limiting plugin, if present
	for _, group := range targetContent.ConsumerGroups {
		for _, plugin := range group.Plugins {
			// 2. If it's missing the namespace
			if (*plugin.Name == RateLimitingPluginName) && (plugin.Config["namespace"] == nil) {
				fmt.Printf("Found rate-limiting plugin in consumer group %s\n", *group.Name)

				// 3. Find the nearest plugin (from route-scope, backwards) that reference
				//    this consumer group in its overrides
				relatedPlugin := findNearestPlugin(group.Name, targetContent)

				if relatedPlugin == nil {
					fmt.Printf("No related plugin found for consumer group %s\n", *group.Name)
				} else {
					fmt.Printf("Found related plugin %s\n", *relatedPlugin.Name)

					// 4. Replace the consumer group's fake plugin with a deepcopy of the related one
					replacementLimit := plugin.Config["limit"]
					replacementWindowSize := plugin.Config["window_size"]
					replacementRetryAfterJitterMax := plugin.Config["retry_after_jitter_max"]
					replacementWindowType := plugin.Config["window_type"]

					plugin.Config = relatedPlugin.Config.DeepCopy()

					// 5. Now override the consumer_group specific config options
					plugin.Config["limit"] = replacementLimit
					plugin.Config["window_size"] = replacementWindowSize
					plugin.Config["retry_after_jitter_max"] = replacementRetryAfterJitterMax
					plugin.Config["window_type"] = replacementWindowType
					plugin.Config["consumer_groups"] = nil
					plugin.Config["enforce_consumer_groups"] = false
				}
			}
		}
	}

	// 6. Finally, remove all policy references from all rate-limiting-advanced plugins
	//    because we don't need them anymore
	removeAllPluginConsumerGroups(targetContent)

	// 7. Serialize the Kong state object back to a file
	err = file.WriteContentToFile(targetContent, destination, file.Format("YAML"))
}
