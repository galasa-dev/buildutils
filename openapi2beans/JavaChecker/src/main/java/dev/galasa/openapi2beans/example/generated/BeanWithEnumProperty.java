package dev.galasa.openapi2beans.example.generated;

// bean with an enum property
public class BeanWithEnumProperty {
    // Class Variables //
    // an enum with 2 values to test against.
    private AnEnumProperty anEnumProperty;

    // Constants //

    public BeanWithEnumProperty (AnEnumProperty anEnumProperty) {
        this.anEnumProperty = anEnumProperty;
    }

    // Getters //
    public AnEnumProperty GetAnEnumProperty() {
        return this.anEnumProperty;
    }

    // Setters //
    public void SetAnEnumProperty(AnEnumProperty anEnumProperty) {
        this.anEnumProperty = anEnumProperty;
    }
}