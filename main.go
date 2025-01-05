package main

import (
	"fmt"
	"log"

	"github.com/tiggoins/harbor-lister/config"
	"github.com/tiggoins/harbor-lister/services"
)

func main() {
	config := config.ParseFlags()

	lister := services.NewHarborLister(config)
	if err := lister.List(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("列表生成完成，数据已保存到:", config.OutputFile)
}
