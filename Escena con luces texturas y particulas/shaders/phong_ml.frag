#version 410 core
out vec4 FragColor;


struct PointLight {
    vec3 position;
    
    float constant;
    float linear;
    float quadratic;
	vec3 lightColor;
    vec3 ambient;
    vec3 diffuse;
    vec3 specular;
};



#define NR_POINT_LIGHTS 8

in vec3 FragPos;
in vec3 Normal;
in vec2 TexCoord;

uniform int numLights;
uniform vec3 objectColor;
uniform vec3 viewPos;
uniform sampler2D texSampler0;
uniform sampler2D texSampler1;
uniform PointLight pointLights[NR_POINT_LIGHTS];


// function prototypes

vec3 CalcPointLight(PointLight light, vec3 normal, vec3 fragPos, vec3 viewDir);


void main()
{    
    // properties
    vec3 norm = normalize(Normal);
    vec3 viewDir = normalize(viewPos - FragPos);
    
    // == =====================================================
    // Our lighting is set up in 3 phases: directional, point lights and an optional flashlight
    // For each phase, a calculate function is defined that calculates the corresponding color
    // per lamp. In the main() function we take all the calculated colors and sum them up for
    // this fragment's final color.
    // == =====================================================
    // phase 1: directional lighting
    vec3 result = vec3(0.0,0.0,0.0);
    // phase 2: point lights
    for(int i = 0; i < numLights; i++)
        result += CalcPointLight(pointLights[i], norm, FragPos, viewDir);    
    // phase 3: spot light
   // result += CalcSpotLight(spotLight, norm, FragPos, viewDir);
    result = result * objectColor;
    if (textureSize(texSampler0, 0).x > 1){
        if (textureSize(texSampler1, 0).x > 1){
            FragColor = mix(texture(texSampler1, TexCoord), texture(texSampler0, TexCoord), 0.5) * vec4(result, 1.0f);
        }else {
            FragColor = texture(texSampler0, TexCoord) * vec4(result, 1.0);
        }
    }
    else {
        FragColor = vec4(result, 1.0);
    }    
    
}


// calculates the color when using a point light.
vec3 CalcPointLight(PointLight light, vec3 normal, vec3 fragPos, vec3 viewDir)
{
    vec3 lightDir = normalize(light.position - fragPos);
    // diffuse shading
    float diff = max(dot(normal, lightDir), 0.0);
    // specular shading
    vec3 reflectDir = reflect(-lightDir, normal);
    float spec = pow(max(dot(viewDir, reflectDir), 0.0), 32.0);
    // attenuation
    float pdistance = length(light.position - fragPos);
    float attenuation = 1.0 / (light.constant + light.linear * pdistance + light.quadratic * (pdistance * pdistance));    
    // combine results
    vec3 ambient = light.ambient;
    vec3 diffuse = light.diffuse * diff * light.lightColor;
    vec3 specular = light.specular * spec * light.lightColor;
    ambient *= attenuation;
    diffuse *= attenuation;
    specular *= attenuation;
    return (ambient + diffuse + specular);
}