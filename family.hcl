local {
  date = "19 Jan 2024"
}

version = 1

// description for elahe
person "Elahe Dastan" {
  birthday = 1999
  date = local.date
}

# description for parham
person "Parham Alvani" {
  birthday = 1995
  date = elahe_dastan.date
}
