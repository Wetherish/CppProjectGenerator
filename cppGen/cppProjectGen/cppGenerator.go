package cppGenerator

import (
	"fmt"
	"os"
)

func GenerateCppProject(path string, projectName string) error {
	err := os.Chdir(path)
	fmt.Print("DONE " + projectName)
	if err != nil {
		return err
	}

	err = generateProjectStructure(projectName)
	if err != nil {
		return err
	}
	fmt.Print("DONE generateProjectStructure")
	err = createCmakeList(projectName)
	if err != nil {
		return err
	}

	fmt.Print("DONE generateProjectStructure")
	err = generateTests()
	if err != nil {
		return err
	}
	fmt.Print("DONE generateProjectStructure")
	err = createLibs()
	if err != nil {
		return err
	}
	fmt.Print("DONE generateProjectStructure")
	err = createMain()
	if err != nil {
		return err
	}

	return nil
}

func generateProjectStructure(projectname string) error {
	err := os.Mkdir(projectname, os.ModePerm)
	if err != nil {
		return err
	}
	err = os.Chdir(projectname)
	if err != nil {
		return err
	}
	directories := []string{"build", "src", "libs", "tests"}

	for _, dir := range directories {
		err := os.Mkdir(dir, os.ModePerm)
		if err != nil {
			return err
		}
	}
	return nil
}
func createCmakeList(Project string) error {
	cmakeFile, err := os.Create("CMakeLists.txt")
	if err != nil {
		return err
	}

	content := `cmake_minimum_required(VERSION 3.10)
	project(` + Project + `)
	set(CMAKE_CXX_STANDARD 20)
	set(CMAKE_EXPORT_COMPILE_COMMANDS ON)
	add_executable(main main.cpp)
	target_include_directories(main PUBLIC libs)
	target_link_libraries(main modern_library)
	include(FetchContent)
	FetchContent_Declare(
		googletest
		URL https://github.com/google/googletest/archive/03597a01ee50ed33e9dfd640b249b4be3799d395.zip
	)
	FetchContent_MakeAvailable(googletest)
	add_subdirectory(tests)
	add_subdirectory(libs)
	`

	_, err = cmakeFile.WriteString(content)
	if err != nil {
		return err
	}

	defer cmakeFile.Close()
	return nil
}

func generateTests() error {
	err := os.Chdir("tests")
	if err != nil {
		return err
	}

	err = testCMake()
	if err != nil {
		return err
	}

	testFile, err := os.Create("test.cpp")
	if err != nil {
		return err
	}

	_, err = testFile.WriteString(`#include <gtest/gtest.h>

TEST(ExampleTest, ExampleTest1) {
    EXPECT_EQ(1, 1);
}
`)
	if err != nil {
		return err
	}
	defer testFile.Close()
	err = os.Chdir("..")
	if err != nil {
		return err
	}
	return nil
}

func testCMake() error {

	cmakeFile, err := os.Create("CMakeLists.txt")
	if err != nil {
		return err
	}
	content := `add_executable(hello_test test.cpp)
	target_link_libraries(hello_test GTest::gtest_main)
	include(GoogleTest)
	gtest_discover_tests(hello_test)
	`
	_, err = cmakeFile.WriteString(content)
	if err != nil {
		return err
	}
	defer cmakeFile.Close()
	return nil
}

func createMain() error {
	mainFile, err := os.Create("main.cpp")
	if err != nil {
		return err
	}
	content := `#include <class.hpp>
	#include <iostream>
	
	int main() {
	  MyClass myClass;
	  myClass.setNumber(5);
	  myClass.multiplyNumber();
	  std::cout << myClass.getNumber() << std::endl;
	
	  return 0;
	}
	`
	_, err = mainFile.WriteString(content)
	if err != nil {
		return err
	}
	return nil
}

func createClass() error {
	classFile, err := os.Create("class.hpp")
	if err != nil {
		return err
	}
	content := `#ifndef CLASS_HPP
	#define CLASS_HPP
	
	class MyClass
	{
	public:
		MyClass() = default; 
		~MyClass() = default; 
	
		int getNumber() const;
		void setNumber(int number);
		void multiplyNumber();
	
	private:
		const int constNumber = 3;
		int number2;
	};
	
	#endif // CLASS_HPP
	`
	_, err = classFile.WriteString(content)
	if err != nil {
		return err
	}
	return nil
}

func createClassCpp() error {
	classFile, err := os.Create("class.cpp")
	if err != nil {
		return err
	}
	content := `#include "class.hpp"
	int MyClass::getNumber() const
	{
		return number2;
	}
	void MyClass::setNumber(int number)
	{
		number2 = number;
	}
	void MyClass::multiplyNumber()
	{
		number2 *= constNumber;
	}
	`
	_, err = classFile.WriteString(content)
	if err != nil {
		return err
	}
	return nil
}
func createCmakeListLib() error {

	cmakeFile, err := os.Create("CMakeLists.txt")
	if err != nil {
		return err
	}
	content := `add_library(modern_library STATIC)
	target_sources(modern_library PRIVATE class.cpp)
	target_include_directories(modern_library PUBLIC ${CMAKE_CURRENT_SOURCE_DIR})
	`
	_, err = cmakeFile.WriteString(content)
	if err != nil {
		return err
	}
	defer cmakeFile.Close()
	return nil
}

func createLibs() error {
	err := os.Chdir("libs")
	if err != nil {
		return err
	}
	err = createClass()
	if err != nil {
		return err
	}
	err = createClassCpp()
	if err != nil {
		return err
	}
	err = createCmakeListLib()
	if err != nil {
		return err
	}
	err = os.Chdir("..")
	if err != nil {
		return err
	}
	return nil
}
