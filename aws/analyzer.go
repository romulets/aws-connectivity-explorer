package aws

import (
	"asset-relations/aws/awsfetcher"
	"encoding/json"
	"log/slog"
	"os"
)

type analyzer struct {
	logger *slog.Logger
}

func newAnalyzer(logger *slog.Logger) analyzer {
	return analyzer{logger: logger}
}

//func (a *analyzer) analyze(ec2Instances []awsfetcher.Ec2Instance) {
//	data, err := json.MarshalIndent(ec2Instances, "", "  ")
//	if err != nil {
//		return
//	}
//
//	file, errs := os.Create("debug.json")
//	if errs != nil {
//		a.logger.Error("Failed to create file:", errs)
//		return
//	}
//	defer file.Close()
//
//	_, errs = file.Write(data)
//	if errs != nil {
//		a.logger.Error("Failed to write to file:", errs) //print the failed message
//		return
//	}
//}

func (a *analyzer) analyze(ec2Instances []awsfetcher.Ec2Instance) {
	data, err := json.MarshalIndent(groupInstancesByVPC(ec2Instances), "", "  ")
	if err != nil {
		return
	}

	file, errs := os.Create("debug.json")
	if errs != nil {
		a.logger.Error("Failed to create file:", errs)
		return
	}
	defer file.Close()

	_, errs = file.Write(data)
	if errs != nil {
		a.logger.Error("Failed to write to file:", errs) //print the failed message
		return
	}
}

func groupInstancesByVPC(instances []awsfetcher.Ec2Instance) map[string][]awsfetcher.Ec2Instance {
	grouped := make(map[string][]awsfetcher.Ec2Instance, len(instances))

	for _, inst := range instances {
		group, exists := grouped[inst.VPC]
		if !exists {
			group = make([]awsfetcher.Ec2Instance, 0, len(instances))
		}

		group = append(group, inst)
		grouped[inst.VPC] = group
	}

	return grouped
}
