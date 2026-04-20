#!/usr/bin/env python3
"""
Generate movie posters using Google Imagen 3 via Vertex AI.
Each poster is a photorealistic cinematic image matching the style of the
existing Okteto movie posters: dramatic lighting, actor name puns at top,
title at bottom, "An Okteto Original" badge.
"""

import os
import time
import io

import vertexai
from vertexai.preview.vision_models import ImageGenerationModel
from PIL import Image, ImageDraw, ImageFont

PROJECT_ID = "vast-watch-300207"
LOCATION = "us-central1"
OUT = "/Users/ramiro/okteto/movies-catalog-20/frontend/src/static"
W, H = 400, 600
FONT_PATH = "/System/Library/Fonts/HelveticaNeue.ttc"


def font(size, idx=0):
    return ImageFont.truetype(FONT_PATH, size, index=idx)


def draw_centered(draw, text, y, fnt, color, shadow=(0, 0, 0)):
    bb = fnt.getbbox(text)
    tw = bb[2] - bb[0]
    x = (W - tw) // 2
    for dx, dy in [(-2, -2), (2, -2), (-2, 2), (2, 2), (0, 3), (3, 0), (-3, 0)]:
        draw.text((x + dx, y + dy), text, fill=shadow + (160,), font=fnt)
    draw.text((x, y), text, fill=color, font=fnt)


def best_font_size(lines, max_w, max_size=80, min_size=28):
    for size in range(max_size, min_size - 1, -2):
        f = font(size, 9)
        try:
            if all(f.getbbox(l)[2] - f.getbbox(l)[0] <= max_w for l in lines):
                return f
        except Exception:
            pass
    return font(min_size, 9)


def overlay_text(img_bytes, title_lines, actors, accent_rgb):
    """Overlay title, actors, and badge onto the Imagen-generated image."""
    img = Image.open(io.BytesIO(img_bytes)).convert("RGB").resize((W, H), Image.LANCZOS)

    # Bottom gradient for title readability
    from PIL import ImageDraw as ID
    grad = Image.new("RGBA", (W, H), (0, 0, 0, 0))
    d = ID.Draw(grad)
    start = H // 2
    for y in range(start, H):
        t = (y - start) / (H - start)
        alpha = int(t ** 1.3 * 200)
        d.line([(0, y), (W, y)], fill=(0, 0, 0, alpha))
    img = Image.alpha_composite(img.convert("RGBA"), grad).convert("RGB")

    # Top bar for actor names
    bar = Image.new("RGBA", (W, H), (0, 0, 0, 0))
    d2 = ID.Draw(bar)
    for y in range(55):
        a = int(150 * (1 - y / 55) ** 0.6)
        d2.line([(0, y), (W, y)], fill=(0, 0, 0, a))
    img = Image.alpha_composite(img.convert("RGBA"), bar).convert("RGB")

    draw = ImageDraw.Draw(img)

    # Title
    title_fnt = best_font_size(title_lines, W - 24)
    total_h = sum(title_fnt.getbbox(l)[3] - title_fnt.getbbox(l)[1] + 10 for l in title_lines)
    y = H - total_h - 44
    for line in title_lines:
        bb = title_fnt.getbbox(line)
        lh = bb[3] - bb[1]
        draw_centered(draw, line, y, title_fnt, accent_rgb)
        y += lh + 10

    # Actors
    draw_centered(draw, actors, 14, font(13, 0), (230, 230, 230))

    # Badge
    badge_fnt = font(12, 0)
    badge = "An  okteto  Original"
    bb = badge_fnt.getbbox(badge)
    tw = bb[2] - bb[0]
    bx = (W - tw) // 2
    draw.text((bx + 1, H - 26 + 1), badge, fill=(0, 0, 0, 130), font=badge_fnt)
    draw.text((bx, H - 26), badge, fill=(255, 255, 255, 210), font=badge_fnt)

    return img


MOVIES = [
    dict(
        filename="poster-thesidecar.png",
        title_lines=["THE", "SIDECAR"],
        actors="JACK NICHOLSCALE   SHELLEY DUVNAMESPACE",
        accent=(255, 255, 255),
        prompt=(
            "Cinematic horror movie poster photograph. A grand, decaying Victorian hotel ballroom "
            "at night, empty except for a lone figure at the end of a long dark corridor. "
            "Frozen winter outside tall windows. Cold blue and white tones with deep shadows. "
            "Photorealistic, dramatic film lighting, moody atmosphere, unsettling."
        ),
    ),
    dict(
        filename="poster-podfiction.png",
        title_lines=["POD", "FICTION"],
        actors="JOHN TRAVOLTA-NODE   UMA THURCONTAINER",
        accent=(255, 210, 0),
        prompt=(
            "Cinematic crime thriller movie poster photograph. A 1990s American diner interior, "
            "two stylishly dressed figures in suits sitting in a booth, one pointing dramatically. "
            "Warm amber and golden lighting, neon signs, cigarette smoke haze. "
            "Photorealistic, high contrast, gritty film noir aesthetic."
        ),
    ),
    dict(
        filename="poster-eternalstateless.png",
        title_lines=["ETERNAL", "STATELESS"],
        actors="JIM CARREY-ON   KATE WINSOCKET",
        accent=(180, 220, 255),
        prompt=(
            "Cinematic romantic drama movie poster photograph. Two figures lying in snow on a frozen "
            "lake at night, surrounded by aurora-like streaks of light. Dreamlike, memories fading. "
            "Cold blue and soft lavender tones, ethereal fog. "
            "Photorealistic, melancholic, visually stunning."
        ),
    ),
    dict(
        filename="poster-thereplica.png",
        title_lines=["THE", "REPLICA"],
        actors="LEONARD DICAPICHAT   TOM HARDYAML",
        accent=(200, 230, 255),
        prompt=(
            "Cinematic survival epic movie poster photograph. A lone figure crawling through a frozen "
            "wilderness, towering snow-capped mountains and pine forests behind them, bear tracks in snow. "
            "Pale cold winter light, breath visible, extreme isolation. "
            "Photorealistic, raw and brutal, widescreen composition."
        ),
    ),
    dict(
        filename="poster-ingress.png",
        title_lines=["INGRESS"],
        actors="LEONARDO DICAPICHAT   JOSEPH GORDON-LEAVITT-LOAD",
        accent=(0, 210, 180),
        prompt=(
            "Cinematic sci-fi thriller movie poster photograph. A city street folding and rotating "
            "in on itself defying gravity, buildings bending at impossible angles, figures in suits "
            "walking sideways. Teal and grey tones, dramatic perspective. "
            "Photorealistic, mind-bending architecture, Christopher Nolan style."
        ),
    ),
    dict(
        filename="poster-darknamespace.png",
        title_lines=["THE DARK", "NAMESPACE"],
        actors="CHRISTIAN KUBELE   HEATH LEDGER-NODE",
        accent=(255, 220, 60),
        prompt=(
            "Cinematic superhero thriller movie poster photograph. A dark city skyline at night "
            "with a single bright spotlight cutting through storm clouds forming a symbol. "
            "A lone caped figure stands at the edge of a skyscraper rooftop. "
            "Dark navy and charcoal tones with golden highlights. Photorealistic, epic scale."
        ),
    ),
    dict(
        filename="poster-tainttoleration.png",
        title_lines=["TAINT &", "TOLERATION"],
        actors="KEIRA KNIGHTLEY-LOAD   MATTHEW MACFADYEN-SET",
        accent=(255, 240, 190),
        prompt=(
            "Cinematic period romance movie poster photograph. An English country estate at golden hour, "
            "a woman in a Regency-era dress standing in a flower meadow, a gentleman approaching "
            "across rolling green hills. Warm amber and ivory tones, soft diffused sunlight. "
            "Photorealistic, painterly quality, Jane Austen aesthetic."
        ),
    ),
    dict(
        filename="poster-missiondeployable.png",
        title_lines=["MISSION:", "DEPLOYABLE"],
        actors="TOM CRONJOB",
        accent=(255, 130, 30),
        prompt=(
            "Cinematic action thriller movie poster photograph. A lone figure in tactical gear "
            "running across a rooftop at dusk, city lights below, explosion behind them. "
            "Deep orange and red sunset tones, motion blur, smoke trails. "
            "Photorealistic, high-octane action, cinematic scale."
        ),
    ),
    dict(
        filename="poster-thehelm.png",
        title_lines=["THE", "HELM"],
        actors="MARTIN SHEEN-SERVER   MARLON BRANDONODE",
        accent=(220, 160, 40),
        prompt=(
            "Cinematic war epic movie poster photograph. A military patrol boat moving through "
            "a dense jungle river at dusk, soldiers silhouetted against a burning orange sky, "
            "mist rising from dark water. Apocalyptic orange and amber tones, heavy atmosphere. "
            "Photorealistic, Vietnam War era aesthetic, foreboding."
        ),
    ),
    dict(
        filename="poster-schindlersregistry.png",
        title_lines=["SCHINDLER'S", "REGISTRY"],
        actors="LIAM NEESON-NODE   RALPH FIENNES-LOG",
        accent=(220, 50, 50),
        prompt=(
            "Cinematic historical drama movie poster photograph. A black and white scene of a "
            "crowded wartime factory, one small red coat visible among grey figures. "
            "Stark contrast, heavy shadows, single point of color drawing the eye. "
            "Photorealistic, somber and powerful, Spielberg aesthetic."
        ),
    ),
    dict(
        filename="poster-nodesofwrath.png",
        title_lines=["NODES", "OF WRATH"],
        actors="HENRY FONDAMENTALS   JANE FONDATION",
        accent=(230, 170, 60),
        prompt=(
            "Cinematic Depression-era drama movie poster photograph. A weathered family on a dusty "
            "road beside an overloaded truck, vast Oklahoma dust bowl plains stretching to the horizon. "
            "Sepia and amber tones, harsh sunlight, determination and hardship. "
            "Photorealistic, 1930s Americana, stark and dignified."
        ),
    ),
    dict(
        filename="poster-resourcequota.png",
        title_lines=["THE RESOURCE", "QUOTA"],
        actors="STEVE MCQUEUARRY   JAMES GARTNER",
        accent=(200, 220, 140),
        prompt=(
            "Cinematic WWII adventure movie poster photograph. A group of soldiers in a POW camp "
            "digging a tunnel under barbed wire fences, searchlights sweeping the yard at night. "
            "Cool moonlit blues and greens, tense atmosphere. "
            "Photorealistic, classic war film aesthetic, Steve McQueen era."
        ),
    ),
    dict(
        filename="poster-anamespace.png",
        title_lines=["A NAMESPACE", "OF THEIR OWN"],
        actors="GEENA DAEMONSET   MADONNA-NODE",
        accent=(100, 240, 160),
        prompt=(
            "Cinematic sports drama movie poster photograph. A women's baseball team celebrating "
            "on a sunlit field in 1940s uniforms, dugout and stadium crowd behind them, summer light. "
            "Warm golden afternoon tones, joy and camaraderie. "
            "Photorealistic, nostalgic Americana, uplifting."
        ),
    ),
    dict(
        filename="poster-persistentvolume.png",
        title_lines=["THE PERSISTENT", "VOLUME"],
        actors="TIM ROBBINS-HOOK   MORGAN FREECONTAINER",
        accent=(130, 180, 240),
        prompt=(
            "Cinematic prison drama movie poster photograph. A man standing in a prison yard "
            "in pouring rain, arms outstretched looking up at the sky in liberation, stone walls "
            "behind him. Cool blue and grey tones, dramatic rain, freedom. "
            "Photorealistic, emotionally powerful, Shawshank aesthetic."
        ),
    ),
]


def generate_poster(movie, model):
    filename = movie["filename"]
    out_path = os.path.join(OUT, filename)

    print(f"\n{'─'*50}")
    print(f"  {filename}")

    try:
        print(f"  → Generating with Imagen 3...")
        response = model.generate_images(
            prompt=movie["prompt"],
            number_of_images=1,
            aspect_ratio="9:16",
            safety_filter_level="block_few",
            person_generation="allow_adult",
        )
        img_bytes = response.images[0]._image_bytes
        print(f"  ✓ Generated")
    except Exception as e:
        print(f"  ✗ Generation failed: {e}")
        return False

    try:
        img = overlay_text(img_bytes, movie["title_lines"], movie["actors"], movie["accent"])
        img.save(out_path)
        print(f"  ✓ Saved → {filename}")
        return True
    except Exception as e:
        print(f"  ✗ Overlay failed: {e}")
        return False


if __name__ == "__main__":
    print(f"Initializing Vertex AI (project={PROJECT_ID}, location={LOCATION})...")
    vertexai.init(project=PROJECT_ID, location=LOCATION)
    model = ImageGenerationModel.from_pretrained("imagen-3.0-generate-001")

    print(f"Generating {len(MOVIES)} posters → {OUT}\n")
    ok, fail = 0, 0
    for movie in MOVIES:
        success = generate_poster(movie, model)
        if success:
            ok += 1
        else:
            fail += 1
        time.sleep(2)  # rate limiting

    print(f"\n{'='*50}")
    print(f"Done: {ok} succeeded, {fail} failed")
