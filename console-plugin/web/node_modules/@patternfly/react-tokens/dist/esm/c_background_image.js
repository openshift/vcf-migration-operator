export const c_background_image = {
  ".pf-v5-c-background-image": {
    "c_background_image_BackgroundColor": {
      "name": "--pf-v5-c-background-image--BackgroundColor",
      "value": "#151515",
      "values": [
        "--pf-v5-global--BackgroundColor--dark-100",
        "$pf-v5-global--BackgroundColor--dark-100",
        "$pf-v5-color-black-900",
        "#151515"
      ]
    },
    "c_background_image_BackgroundImage": {
      "name": "--pf-v5-c-background-image--BackgroundImage",
      "value": "none"
    },
    "c_background_image_BackgroundSize_min_width": {
      "name": "--pf-v5-c-background-image--BackgroundSize--min-width",
      "value": "200px"
    },
    "c_background_image_BackgroundSize_width": {
      "name": "--pf-v5-c-background-image--BackgroundSize--width",
      "value": "60%"
    },
    "c_background_image_BackgroundSize_max_width": {
      "name": "--pf-v5-c-background-image--BackgroundSize--max-width",
      "value": "600px"
    },
    "c_background_image_BackgroundSize": {
      "name": "--pf-v5-c-background-image--BackgroundSize",
      "value": "clamp(200px, 60%, 600px)",
      "values": [
        "clamp(--pf-v5-c-background-image--BackgroundSize--min-width, --pf-v5-c-background-image--BackgroundSize--width, --pf-v5-c-background-image--BackgroundSize--max-width)",
        "clamp(200px, 60%, 600px)"
      ]
    },
    "c_background_image_BackgroundPosition": {
      "name": "--pf-v5-c-background-image--BackgroundPosition",
      "value": "bottom right"
    }
  },
  ":where(.pf-v5-m-dir-rtl, [dir=rtl]) .pf-v5-c-background-image": {
    "c_background_image_BackgroundPosition": {
      "name": "--pf-v5-c-background-image--BackgroundPosition",
      "value": "bottom left"
    }
  }
};
export default c_background_image;