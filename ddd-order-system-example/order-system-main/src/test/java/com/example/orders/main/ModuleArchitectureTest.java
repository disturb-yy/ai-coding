package com.example.orders.main;

import static com.tngtech.archunit.lang.syntax.ArchRuleDefinition.noClasses;

import com.tngtech.archunit.junit.AnalyzeClasses;
import com.tngtech.archunit.junit.ArchTest;
import com.tngtech.archunit.lang.ArchRule;
import com.tngtech.archunit.core.importer.ImportOption;

/** 防止模块规则只停留在文档中。 */
@AnalyzeClasses(
        packages = "com.example.orders",
        importOptions = ImportOption.DoNotIncludeTests.class)
class ModuleArchitectureTest {

    @ArchTest
    static final ArchRule domain_does_not_depend_on_outer_layers = noClasses()
            .that().resideInAnyPackage("..domain..")
            .should().dependOnClassesThat().resideInAnyPackage("..api..", "..app..", "..base..", "..main..");

    @ArchTest
    static final ArchRule api_does_not_depend_on_implementations = noClasses()
            .that().resideInAnyPackage("..api..")
            .should().dependOnClassesThat().resideInAnyPackage("..domain..", "..app..", "..base..", "..main..");

    @ArchTest
    static final ArchRule app_does_not_depend_on_infrastructure = noClasses()
            .that().resideInAnyPackage("..app..")
            .should().dependOnClassesThat().resideInAnyPackage("..base..", "..main..");

    @ArchTest
    static final ArchRule controllers_depend_on_api_not_app_implementation = noClasses()
            .that().resideInAPackage("..main..")
            .and().haveSimpleNameEndingWith("Controller")
            .should().dependOnClassesThat().resideInAPackage("..app..");
}
